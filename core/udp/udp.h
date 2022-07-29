#ifndef __OPENUDP_UDP_H
#define __OPENUDP_UDP_H  1

#include <netinet/in.h>

#include "openvswitch/shash.h"

struct udp_message {
    u_int32_t padding[2];
    u_int32_t spi;
    u_int32_t seqno;
};

struct udp_server {
    u_int16_t port;
    int32_t socket;
    long long int send_t;
};

struct udp_connect {
    int32_t socket;
    int32_t remote_port;
    const char *remote_address;
    u_int32_t spi;
    u_int32_t seqno;
};

int send_ping_once(struct udp_connect *);
int recv_ping_once(struct udp_server *, struct sockaddr_in *, u_int8_t *, size_t);

int open_socket(struct udp_server *);
int configure_socket(struct udp_server *);

static inline void shash_empty(struct shash *sh)
{
    shash_destroy(sh);
    shash_init(sh);
}

#endif
