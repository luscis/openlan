/*
 * Copyright (c) 2021-2022 OpenLAN Inc.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 */

#ifndef CORE_SOCKET_H
#define CORE_SOCKET_H

#include "types.h"

typedef struct {
    int socket_fd;
    int device_fd;
} peer_t;

int start_tcp_server(uint16_t port);
int start_tcp_client(const char *addr, uint16_t port);

#endif //CORE_SOCKET_H
