/*
 * Copyright (c) 2021-2022 OpenLAN Inc.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 */

#include <config.h>
#include <errno.h>
#include <getopt.h>
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#include "openvswitch/dynamic-string.h"
#include "openvswitch/poll-loop.h"
#include "openvswitch/vconn.h"
#include "openvswitch/vlog.h"

#include "ovsdb-data.h"
#include "ovsdb-idl-provider.h"

#include "command-line.h"
#include "confd-idl.h"
#include "daemon.h"
#include "udp.h"
#include "unixctl.h"
#include "ovs-thread.h"
#include "timeval.h"
#include "version.h"

#define RUN_DIR   "/var/openlan"
#define UDP_PORT  4500

VLOG_DEFINE_THIS_MODULE(main);
/* Rate limit for error messages. */
static struct vlog_rate_limit rl = VLOG_RATE_LIMIT_INIT(5, 5);

static char *default_db_ = NULL;
static char *db_remote = NULL;
static int32_t udp_port = 0;

struct udp_context {
    struct udp_server *srv;
    struct ovsdb_idl *idl;
    struct ovsdb_idl_txn *idl_txn;
    struct shash names;
    struct shash networks;
    struct shash links;

};

static inline char *
unixctl_dir()
{
    return xasprintf("%s/%s.ctl", RUN_DIR, program_name);
}

static inline char *
default_db(void)
{
    if (!default_db_) {
        default_db_ = xasprintf("unix:%s/confd.sock", RUN_DIR);
    }
    return default_db_;
}

static void
usage(void)
{
    printf("\
%s: OpenLAN UDP Connection\n\
usage %s [OPTIONS]\n\
\n\
Options:\n\
  --port=PORT             listen on local udp PORT\n\
                          (default: %d)\n\
  --db=DATABASE           connect to database at DATABASE\n\
                          (default: %s)\n\
  -h, --help              display this help message\n\
  -o, --options           list available options\n\
  -V, --version           display version information\n\
", program_name, program_name, UDP_PORT, default_db());
    vlog_usage();
    exit(EXIT_SUCCESS);
}

static void
parse_options(int argc, char *argv[])
{
    enum {
        VLOG_OPTION_ENUMS,
    };

    static struct option long_options[] = {
        {"port", required_argument, NULL, 'p'},
        {"db", required_argument, NULL, 'd'},
        {"help", no_argument, NULL, 'h'},
        {"version", no_argument, NULL, 'V'},
        VLOG_LONG_OPTIONS,
        {NULL, 0, NULL, 0}
    };
    char *short_options = ovs_cmdl_long_options_to_short_options(long_options);

    for (;;) {
        int c;

        c = getopt_long(argc, argv, short_options, long_options, NULL);
        if (c == -1) {
            break;
        }

        switch (c) {
        case 'd':
            db_remote = xstrdup(optarg);
            break;

        case 'p':
            udp_port = atoi(optarg);
            break;

        case 'h':
            usage();

        case 'V':
            ovs_print_version(OFP13_VERSION, OFP13_VERSION);
            exit(EXIT_SUCCESS);

        VLOG_OPTION_HANDLERS

        case '?':
            exit(EXIT_FAILURE);

        default:
            abort();
        }
    }
    free(short_options);

    if (!db_remote) {
        db_remote = xstrdup(default_db());
    }
    if (!udp_port) {
        udp_port = UDP_PORT;
    }
}

static void
udp_exit(struct unixctl_conn *conn, int argc OVS_UNUSED,
        const char *argv[] OVS_UNUSED, void *exiting_)
{
    bool *exiting = exiting_;
    *exiting = true;

    unixctl_command_reply(conn, NULL);
}

static void
cache_run(struct udp_context *ctx)
{
    const struct openrec_name_cache *nc;
    const struct openrec_virtual_network *vn;
    const struct openrec_virtual_link *vl;

    shash_empty(&ctx->names);
    shash_empty(&ctx->networks);
    shash_empty(&ctx->links);

    OPENREC_NAME_CACHE_FOR_EACH (nc, ctx->idl) {
        VLOG_DBG("name_cache: %s %s", nc->name, nc->address);
        shash_add(&ctx->names, nc->name, nc);
    }

    OPENREC_VIRTUAL_NETWORK_FOR_EACH (vn, ctx->idl) {
        VLOG_DBG("virtual_network: %s %s", vn->name, vn->address);
        shash_add(&ctx->networks, vn->name, vn);
    }

    OPENREC_VIRTUAL_LINK_FOR_EACH (vl, ctx->idl) {
        VLOG_DBG("virtual_link: %s %s", vl->network, vl->connection);
        if (!strncmp(vl->connection, "any", 3) || !strlen(vl->connection)) {
            shash_add(&ctx->links, vl->device, vl);
        } else {
            shash_add(&ctx->links, vl->connection, vl);
        }
    }
}

static void
ping_run(struct udp_context *ctx)
{
   char address[128] = {0};
   struct udp_server *srv = ctx->srv;

   if (time_msec() - srv->send_t < 5 *1000) {
      return;
   }

   struct udp_connect conn = {
       .socket = srv->socket,
       .remote_port = UDP_PORT,
       .remote_address = address,
   };
   struct shash_node *node;

   SHASH_FOR_EACH(node, &ctx->links) {
       const struct openrec_virtual_link *vl = node->data;
       if (strncmp(vl->device, "spi:", 4) || strncmp(vl->connection, "udp:", 4)) {
           continue;
       }
       VLOG_DBG("send_ping to %s on %s\n", vl->connection, vl->device);
       ovs_scan(vl->device, "spi:%d", &conn.spi);
       ovs_scan(vl->connection, "udp:%[^:]:%d", address, &conn.remote_port);

       const struct shash_node *nc_node = shash_find(&ctx->names, address);
       if (nc_node) {
           const struct openrec_name_cache *nc = nc_node->data;
           conn.remote_address = nc->address;
       }

       send_ping_once(&conn);
   }
   srv->send_t = time_msec();
}

static void
pong_run(struct udp_context *ctx)
{
    int retval;
    u_int8_t buf[1024];
    struct sockaddr_in from;

    struct udp_server *srv = ctx->srv;
    struct udp_message *data = (struct udp_message *)buf;

    retval = recv_ping_once(srv, &from, buf, sizeof buf);
    if (retval <= 0) {
        return;
    }
    const char *remote_addr = inet_ntoa(from.sin_addr);
    char *spi_conn = xasprintf("spi:%d", ntohl(data->spi));
    struct shash_node *node = shash_find(&ctx->links, spi_conn);

    VLOG_DBG("pong_run from: %s:%d and spi %d\n", remote_addr, ntohs(from.sin_port), ntohl(data->spi));
    if (node) {
        struct openrec_virtual_link *vl = node->data;
        VLOG_DBG("pong_run virtual link: %s %s\n", vl->connection, vl->network);
        struct sockaddr_in dst_addr = from;
        u_int32_t seqno = ntohl(data->seqno) + 1;
        data->seqno = htonl(seqno);
        retval = sendto(srv->socket, data, sizeof *data, 0, (struct sockaddr *)&dst_addr, sizeof dst_addr);
        if (retval <= 0) {
            VLOG_WARN_RL(&rl, "%s: could not send data\n", remote_addr);
        }
        // remote_connection=udp:a.b.c.d:1024
        char *connection = xasprintf("udp:%s:%d", remote_addr, ntohs(from.sin_port));
        openrec_virtual_link_update_status_setkey(vl, "remote_connection", connection);
        free(connection);
    }
    free(spi_conn);
}

static void
ping_wait(struct udp_context *ctx)
{
   poll_timer_wait_until(time_msec() + 5 * 1000);
}

static void
pong_wait(struct udp_context *ctx)
{   
   struct udp_server *srv = ctx->srv;

   poll_fd_wait(srv->socket, POLLIN);
}

int
main(int argc, char *argv[])
{
    struct unixctl_server *unixctl;
    bool exiting = false;
    int retval = 0;
    char *unixdir;

    ovs_cmdl_proctitle_init(argc, argv);
    ovs_set_program_name(argv[0], CORE_PACKAGE_VERSION);

    service_start(&argc, &argv);
    parse_options(argc, argv);

    unixdir = unixctl_dir();
    /* Open and register unixctl */
    retval = unixctl_server_create(unixdir, &unixctl);
    if (retval) {
        goto RET;
    }
    unixctl_command_register("exit", "", 0, 0, udp_exit, &exiting);

    /* Connect to OpenLAN database. */
    struct ovsdb_idl_loop open_idl_loop = OVSDB_IDL_LOOP_INITIALIZER(
        ovsdb_idl_create(db_remote, &openrec_idl_class, true, true));
    ovsdb_idl_get_initial_snapshot(open_idl_loop.idl);

    struct udp_server srv = {
        .port = udp_port,
        .socket = -1,
        .send_t = time_msec(),
    };
    open_socket(&srv);
    if (configure_socket(&srv) < 0) {
        VLOG_ERR("configure_socket: %s\n", strerror(errno));
        goto RET;
    }

    struct udp_context ctx = {
        .idl = open_idl_loop.idl,
        .srv = &srv,
    };

    shash_init(&ctx.names);
    shash_init(&ctx.networks);
    shash_init(&ctx.links);

    while(!exiting) {
        ctx.idl_txn = ovsdb_idl_loop_run(&open_idl_loop);

        if (ctx.idl_txn) {
           cache_run(&ctx);
        }

        ping_run(&ctx);
        pong_run(&ctx);

        ping_wait(&ctx);
        pong_wait(&ctx);

        unixctl_server_run(unixctl);
        unixctl_server_wait(unixctl);
        if (exiting) {
            poll_immediate_wake();
        }
        ovsdb_idl_loop_commit_and_wait(&open_idl_loop);
        poll_block();
        if (should_service_stop()) {
            exiting = true;
        }
    }

    shash_destroy(&ctx.names);
    shash_destroy(&ctx.networks);
    shash_destroy(&ctx.links);

    unixctl_server_destroy(unixctl);
    ovsdb_idl_loop_destroy(&open_idl_loop);
    service_stop();

RET:
    if (db_remote) free(db_remote);
    if (default_db_) free(default_db_);
    if (unixdir) free(unixdir);

    exit(retval);
}
