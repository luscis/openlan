package libol

/*
#include <errno.h>
#include <stdlib.h>
#include <stdio.h>
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

typedef struct {
    u_int32_t padding[2];
    u_int32_t spi;
    u_int32_t seqno;
} udp_message;

typedef struct {
    u_int16_t port;
    int32_t socket;
} udp_server;

typedef struct {
    int32_t socket;
    uint16_t remote_port;
    const char *remote_addr;
    u_int32_t spi;
    u_int32_t seqno;
} udp_connection;

int seqno = 0;

int send_ping_once(udp_connection *conn) {
    int retval = 0;
    struct sockaddr_in dstaddr = {
        .sin_family = AF_INET,
        .sin_port = htons(conn->remote_port),
        .sin_addr = {
            .s_addr = inet_addr(conn->remote_addr),
        },
    };
    udp_message data = {
        .padding = {0, 0},
        .spi = htonl(conn->spi),
    };
    data.seqno = htonl(conn->seqno++);

    retval = sendto(conn->socket, &data, sizeof data, 0, (struct sockaddr *)&dstaddr, sizeof dstaddr);
    if (retval <= 0) {
        printf("%s: could not send data\n", conn->remote_addr);
    }

    return retval;
}

int recv_ping_once(udp_connection *conn,  udp_connection *from) {
    struct sockaddr_in addr;
    int addrlen = sizeof addr;
    udp_message data;
    int datalen = sizeof data;
    int retval = 0;

    memset(&data, 0, sizeof data);
    retval = recvfrom(conn->socket, &data, datalen, 0, (struct sockaddr *)&addr, &addrlen);
    if ( retval <= 0 ) {
        if (errno == EAGAIN) {
            return 0;
        }
        printf("recvfrom: %s\n", strerror(errno));
        return retval;
    }

    from->spi = ntohl(data.spi);
	from->remote_addr = inet_ntoa(addr.sin_addr);
	from->remote_port = ntohs(addr.sin_port);

    return retval;
}

int open_socket(udp_server *srv) {
    int op = 1;
    struct sockaddr_in addr = {
        .sin_family = AF_INET,
        .sin_port = htons(srv->port),
        .sin_addr = {
            .s_addr = INADDR_ANY,
        },
    };
	int retval = 0;

    srv->socket = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
    if (srv->socket == -1) {
        return -1;
    }

	retval = setsockopt(srv->socket, SOL_SOCKET, SO_REUSEADDR, &op, sizeof op);
    if (retval < 0) {
        return -1;
    }
	retval = bind(srv->socket, (struct sockaddr *)&addr, sizeof addr);
    if ( retval == -1) {
        return -1;
    }

    return 0;
}

int configure_socket(udp_server *srv) {
    int encap = UDP_ENCAP_ESPINUDP;

    if (setsockopt(srv->socket, IPPROTO_UDP, UDP_ENCAP, &encap, sizeof encap) < 0) {
        return -1;
    }
    return 0;
}
*/
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

func StartUDP(spi uint32, port uint16, remote string) {
	server := &C.udp_server{
		port: C.ushort(port),
		socket : -1,
	}
	_ = C.open_socket(server)
	C.configure_socket(server)

	addr := C.CString(remote)
	defer C.free(unsafe.Pointer(addr))
	conn := &C.udp_connection {
		socket: server.socket,
		spi: C.uint(spi),
		remote_port: C.ushort(port),
		remote_addr: addr,
	}

	C.send_ping_once(conn)

	w := sync.WaitGroup{}
	w.Add(2)
	Go(func() {
		defer w.Done()
		for i := 0; i < 100; i++ {
			time.Sleep(time.Second)
			C.send_ping_once(conn)
		}
	})

	Go(func() {
		from := &C.udp_connection{}
		C.recv_ping_once(conn, from)
		addr := C.GoString(from.remote_addr)
		fmt.Printf("receive from %s:%d spi %d\n", addr, from.remote_port, from.spi)
		w.Done()
		for {
			from := &C.udp_connection{}
			C.recv_ping_once(conn, from)
			addr := C.GoString(from.remote_addr)
			fmt.Printf("receive from %s:%d spi %d\n", addr, from.remote_port, from.spi)
		}
	})

	w.Wait()
}