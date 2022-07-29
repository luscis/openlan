/*
 * Copyright (c) 2021-2022 OpenLAN Inc.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 */

#include <stdio.h>
#include <string.h>
#include <strings.h>
#include <memory.h>
#include <fcntl.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <assert.h>
#include <pthread.h>

#include "tuntap.h"
#include "socket.h"

int non_blocking(int fd) {
    int flags = fcntl(fd, F_GETFL, 0);
    return fcntl(fd, F_SETFL, flags | O_NONBLOCK);
}

int recv_full(int fd, char *buf, ssize_t size) {
    ssize_t read_size = 0;

    for (; size > 0;) {
        read_size = recv(fd, buf, size, 0);
        if (read_size <= 0) return read_size;
        buf += read_size;
        size -= read_size;
    }
    return 0;
}

int send_full(int fd, char *buf, ssize_t size) {
    ssize_t write_size = 0;

    for (;size > 0;) {
        write_size = send(fd, buf, size, 0);
        if (write_size <= 0) return write_size;
        buf += write_size;
        size -= write_size;
    }
    return 0;
}

void *read_client(void *argv) {
    uint16_t buf_size = 0;
    uint16_t read_size = 0;
    uint8_t buf[4096];
    peer_t *conn = NULL;

    assert(NULL != argv);
    conn = (peer_t *) argv;

    for(;;) {
        buf_size = recv_full(conn->socket_fd, buf, 4);
        if (buf_size != 0) {
            break;
        }
        read_size = ntohs(*(uint16_t *)(buf + 2));
        memset(buf, 0, sizeof buf);
        buf_size = recv_full(conn->socket_fd, buf, read_size);
        if (buf_size != 0) {
            printf("ERROR: on read %d != %d\n", read_size, buf_size);
            break;
        }
        write(conn->device_fd, buf, read_size);
    }
}

void *read_device(void *argv) {
    uint16_t write_size = 0;
    uint16_t read_size = 0;
    uint8_t buf[4096];
    peer_t *conn = NULL;

    assert(NULL != argv);
    conn = (peer_t *) argv;

    for(;;) {
        read_size = read(conn->device_fd, buf + 4, sizeof (buf));
        if (read_size <= 0) {
            continue;
        }
        *(uint16_t *)(buf + 2) = htons(read_size);
        read_size += 4;
        write_size = send_full(conn->socket_fd, buf, read_size);
        if (write_size != 0) {
            printf("ERROR: write to conn %d:%d", write_size, read_size);
            break;
        }
    }
}

int start_peer(peer_t *peer) {
    pthread_t client;
    pthread_t device;

    if(pthread_create(&client, NULL, read_client, &peer)) {
        fprintf(stderr, "Error creating thread\n");
        return 1;
    }
    if(pthread_create(&device, NULL, read_device, &peer)) {
        fprintf(stderr, "Error creating thread\n");
        return 1;
    }
    if(pthread_join(client, NULL)) {
        fprintf(stderr, "Error joining thread\n");
        return 2;
    }
    if(pthread_join(device, NULL)) {
        fprintf(stderr, "Error joining thread\n");
        return 2;
    }
}

int start_tcp_server(uint16_t port) {
    struct sockaddr_in server_addr;

    bzero(&server_addr, sizeof(struct sockaddr_in));
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = htonl(INADDR_ANY);
    server_addr.sin_port = htons(port);

    int server_fd = 0;
    server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if(bind(server_fd, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        printf("bind error\n");
        return -1;
    }
    if(listen(server_fd, 2) < 0) {
        printf("listen error\n");
        return -1;
    }

    struct sockaddr_in conn_addr;
    socklen_t conn_addr_len = sizeof(conn_addr);

    int conn_fd = 0;
    char dev_name[1024] = {0};
    int tap_fd = 0;

    conn_fd = accept(server_fd, (struct sockaddr *)&conn_addr, &conn_addr_len);
    printf("accept connection on %d\n", conn_fd);

    tap_fd = create_tap(dev_name);
    printf("open device on %s with %d\n", dev_name, tap_fd);

    peer_t peer = {
        .socket_fd = conn_fd,
        .device_fd = tap_fd,
    };
   start_peer(&peer);

finish:
    close(conn_fd);
    close(server_fd);
    close(tap_fd);
    printf("exit from %d\n", server_fd);
    return 0;
}

int start_tcp_client(const char *addr, uint16_t port) {
    int ret = 0;
    int socket_fd = 0;
    struct sockaddr_in server_addr;

    socket_fd = socket(PF_INET, SOCK_STREAM, 0);
    if (socket_fd < 0) {
        printf("ERROR: open socket %d", socket_fd);
        return socket_fd;
    }

    bzero(&server_addr, sizeof (server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(port);
    server_addr.sin_addr.s_addr = inet_addr(addr);

    ret = connect(socket_fd, (struct sockaddr *)&server_addr, sizeof(server_addr));
    if(ret ==-1) {
        printf("connect() error\n");
        return ret;
    }
    char dev_name[1024] = {0};
    int tap_fd = 0;

    tap_fd = create_tap(dev_name);
    printf("open device on %s with %d\n", dev_name, tap_fd);

    peer_t peer = {
            .socket_fd = socket_fd,
            .device_fd = tap_fd,
    };
    start_peer(&peer);

finish:
    close(socket_fd);
    close(tap_fd);
    return 0;
}
