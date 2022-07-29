/*
 * Copyright (c) 2021-2022 OpenLAN Inc.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 */

#include <errno.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/ip.h>
#include <linux/udp.h>
#include <linux/xfrm.h>
#include <linux/ipsec.h>
#include <linux/pfkeyv2.h>
#include <arpa/inet.h>

#include "openvswitch/dynamic-string.h"
#include "openvswitch/vlog.h"

#include "socket-util.h"

#include "udp.h"

VLOG_DEFINE_THIS_MODULE(udp);

/* Rate limit for error messages. */
static struct vlog_rate_limit rl = VLOG_RATE_LIMIT_INIT(5, 5);

void
print_hex(const char *prefix, u_int8_t *data, int len)
{
    if (VLOG_IS_DBG_ENABLED()) {
        struct ds s;
        ds_init(&s);
        for (int i = 0; i < len; i++ ) {
            ds_put_format(&s, "%02x ", data[i]);
        }
        VLOG_DBG("%s%s\n", prefix, ds_cstr(&s));
        ds_destroy(&s);
    }
}

int
send_ping_once(struct udp_connect *conn)
{
    int retval = 0;
    struct udp_message data = {
        .padding = {0, 0},
        .spi = htonl(conn->spi),
    };
    data.seqno = htonl(conn->seqno++);
    struct sockaddr_in dst_addr = {
        .sin_family = AF_INET,
        .sin_port = htons(conn->remote_port),
        .sin_addr = {
            .s_addr = inet_addr(conn->remote_address),
        },
    };
    retval = sendto(conn->socket, &data, sizeof data, 0, (struct sockaddr *)&dst_addr, sizeof dst_addr);
    if (retval <= 0) {
        VLOG_WARN_RL(&rl, "%s: could not send data\n", conn->remote_address);
    }
    return retval;
}

int
recv_ping_once(struct udp_server *srv, struct sockaddr_in *addr, u_int8_t *buf, size_t len)
{
    struct udp_message *data = (struct udp_message *)buf;
    int retval = 0, addrlen = sizeof *addr;

    memset(data, 0, sizeof *data);
    retval = recvfrom(srv->socket, buf, len, 0, (struct sockaddr *)addr, &addrlen);
    if ( retval <= 0 ) {
        if (errno == EAGAIN) {
            return 0;
        }
        VLOG_ERR_RL(&rl, "recvfrom: %s\n", strerror(errno));
        return retval;
    }
    const char *remote_addr = inet_ntoa(addr->sin_addr);
    VLOG_DBG("recvfrom: [%s:%d] %d bytes\n", remote_addr, ntohs(addr->sin_port), retval);
    print_hex("recvfrom: ", buf, retval);
    return retval;
}

int
open_socket(struct udp_server *srv)
{
    int op = 1;
    struct sockaddr_in addr = {
        .sin_family = AF_INET,
        .sin_port = htons(srv->port),
        .sin_addr = {
            .s_addr = INADDR_ANY,
        },
    };

    srv->socket = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
    if (srv->socket == -1) {
        return -1;
    }
    if (setsockopt(srv->socket, SOL_SOCKET, SO_REUSEADDR, (void *)&op, sizeof op) < 0) {
        return -1;
    }
    if (bind(srv->socket, (struct sockaddr *)&addr, sizeof addr) == -1) {
        return -1;
    }
    set_nonblocking(srv->socket);

    return srv->socket;
}

int
configure_socket(struct udp_server *srv)
{
    int encap = UDP_ENCAP_ESPINUDP;
    struct xfrm_userpolicy_info pol;

    memset(&pol, 0, sizeof(pol));
    pol.action = XFRM_POLICY_ALLOW;
    pol.sel.family = AF_INET;

    pol.dir = XFRM_POLICY_OUT;
    if (setsockopt(srv->socket, IPPROTO_IP, IP_XFRM_POLICY, &pol, sizeof pol) < 0) {
        return -1;
    }
    pol.dir = XFRM_POLICY_IN;
    if (setsockopt(srv->socket, IPPROTO_IP, IP_XFRM_POLICY, &pol, sizeof pol) < 0) {
        return -1;
    }
    if (setsockopt(srv->socket, IPPROTO_UDP, UDP_ENCAP, &encap, sizeof encap) < 0) {
        return -1;
    }
    return srv->socket;
}
