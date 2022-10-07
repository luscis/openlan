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
} udpin_message;

typedef struct {
    u_int16_t port;
    int32_t socket;
} udpin_server;

typedef struct {
    int32_t socket;
    uint16_t remote_port;
    const char *remote_addr;
    u_int32_t spi;
    u_int32_t seqno;
} udpin_connection;

int seqno = 0;

int send_ping_once(udpin_connection *conn) {
    int retval = 0;
    struct sockaddr_in dstaddr = {
        .sin_family = AF_INET,
        .sin_port = htons(conn->remote_port),
        .sin_addr = {
            .s_addr = inet_addr(conn->remote_addr),
        },
    };
    udpin_message data = {
        .padding = {0, 0},
        .spi = htonl(conn->spi),
    };
    data.seqno = htonl(conn->seqno);

    retval = sendto(conn->socket, &data, sizeof data, 0, (struct sockaddr *)&dstaddr, sizeof dstaddr);
    return retval;
}

int recv_ping_once(udpin_server *srv,  udpin_connection *from) {
    struct sockaddr_in addr;
    int addrlen = sizeof addr;
    udpin_message data;
    int datalen = sizeof data;
    int retval = 0;

    memset(&data, 0, sizeof data);
    retval = recvfrom(srv->socket, &data, datalen, 0, (struct sockaddr *)&addr, &addrlen);
    if ( retval <= 0 ) {
        if (errno == EAGAIN) {
            return 0;
        }
        return retval;
    }

    from->spi = ntohl(data.spi);
	from->remote_addr = inet_ntoa(addr.sin_addr);
	from->remote_port = ntohs(addr.sin_port);

    return retval;
}

int open_socket(udpin_server *srv) {
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
        return retval;
    }
	retval = bind(srv->socket, (struct sockaddr *)&addr, sizeof addr);
    if ( retval == -1) {
        return retval;
    }

    return 0;
}

int configure_socket(udpin_server *srv) {
    int encap = UDP_ENCAP_ESPINUDP;

	return setsockopt(srv->socket, IPPROTO_UDP, UDP_ENCAP, &encap, sizeof encap);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type UdpInServer struct {
	Port   uint16
	Socket int
	server *C.udpin_server
	SeqNo  uint32
}

type UdpInConnection struct {
	Socket     int
	RemotePort uint16
	RemoteAddr string
	Spi        uint32
}

func (c *UdpInConnection) Connection() string {
	return fmt.Sprintf("%s:%d", c.RemoteAddr, c.RemotePort)
}

func (c *UdpInConnection) String() string {
	return fmt.Sprintf("%d on %s:%d", c.Spi, c.RemoteAddr, c.RemotePort)
}

func (u *UdpInServer) Open() error {
	server := &C.udpin_server{
		port:   C.ushort(u.Port),
		socket: -1,
	}
	if ret := C.open_socket(server); ret < 0 {
		return NewErr("UdpInServer.Open errno:%d", ret)
	}
	if ret := C.configure_socket(server); ret < 0 {
		return NewErr("UdpInServer.Open errno:%d", ret)
	}
	u.server = server
	u.Socket = int(server.socket)
	return nil
}

func (u *UdpInServer) Send(to *UdpInConnection) error {
	u.SeqNo += 1
	addr := C.CString(LookupIP(to.RemoteAddr))
	defer C.free(unsafe.Pointer(addr))
	conn := &C.udpin_connection{
		socket:      u.server.socket,
		spi:         C.uint(to.Spi),
		remote_port: C.ushort(to.RemotePort),
		remote_addr: addr,
		seqno:       C.uint(u.SeqNo),
	}
	if ret := C.send_ping_once(conn); ret < 0 {
		return NewErr("UdpInServer.Ping errno:%d", ret)
	}
	return nil
}

func (u *UdpInServer) Recv() (*UdpInConnection, error) {
	from := &C.udpin_connection{}
	if ret := C.recv_ping_once(u.server, from); ret < 0 {
		return nil, NewErr("UdpInServer.Pong errno:%d", ret)
	}
	return &UdpInConnection{
		RemotePort: uint16(from.remote_port),
		RemoteAddr: C.GoString(from.remote_addr),
		Spi:        uint32(from.spi),
	}, nil
}
