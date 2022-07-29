/*
 * Copyright (c) 2021-2022 OpenLAN Inc.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 */

#ifndef CORE_TUNTAP_H
#define CORE_TUNTAP_H

#include <unistd.h>

#define DEV_NET_TUN "/dev/net/tun"

int create_tap(char *name);

#endif //CORE_TUNTAP_H
