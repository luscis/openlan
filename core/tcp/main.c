/*
 * Copyright (c) 2021-2022 OpenLAN Inc.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 */

#include <stdio.h>
#include "socket.h"

int main(int argc, char *argv[]) {
    char *addr = NULL;
    int port = 9090;

    if (argc > 2) {
        addr = argv[1];
        sscanf(argv[2], "%d", &port);
    } else if (argc > 1) {
        sscanf(argv[1], "%d", &port);
    }

    if (addr == NULL) {
        printf("Listen on %d!\n", port);
        start_tcp_server(port);
    } else {
        printf("Connect to %s:%d\n", addr, port);
        start_tcp_client(addr, port);
    }
    return 0;
}
