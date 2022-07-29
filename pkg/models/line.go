package models

import (
	"net"
	"strconv"
	"time"
)

type Line struct {
	EthType    uint16
	IpSource   net.IP
	IpDest     net.IP
	IpProtocol uint8
	PortDest   uint16
	PortSource uint16
	NewTime    int64
	HitTime    int64
}

func NewLine(t uint16) *Line {
	l := &Line{
		EthType: t,
		NewTime: time.Now().Unix(),
		HitTime: time.Now().Unix(),
	}
	return l
}

func (l *Line) String() string {
	str := strconv.FormatUint(uint64(l.EthType), 10)
	str += ":" + l.IpSource.String()
	str += ":" + l.IpDest.String()
	str += ":" + strconv.FormatUint(uint64(l.IpProtocol), 10)
	str += ":" + strconv.FormatUint(uint64(l.PortSource), 10)
	str += ":" + strconv.FormatUint(uint64(l.PortDest), 10)
	return str
}

func (l *Line) UpTime() int64 {
	return time.Now().Unix() - l.NewTime
}

func (l *Line) LastTime() int64 {
	return time.Now().Unix() - l.HitTime
}
