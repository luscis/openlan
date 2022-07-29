package config

import "github.com/luscis/openlan/pkg/libol"

type Queue struct {
	SockWr int `json:"swr"` // per frames about 1572(1514+4+20+20+14)bytes
	SockRd int `json:"srd"` // per frames
	TapWr  int `json:"twr"` // per frames about 1572((1514+4+20+20+14))bytes
	TapRd  int `json:"trd"` // per frames
	VirSnd int `json:"vsd"`
	VirWrt int `json:"vwr"`
}

var (
	QdSwr = 1024 * 4
	QdSrd = 1024 * 4
	QdTwr = 1024 * 2
	QdTrd = 2
	QdVsd = 1024 * 8
	QdVWr = 1024 * 4
)

func (q *Queue) Default() {
	if q.SockWr == 0 {
		q.SockWr = QdSwr
	}
	if q.SockRd == 0 {
		q.SockRd = QdSrd
	}
	if q.TapWr == 0 {
		q.TapWr = QdTwr
	}
	if q.TapRd == 0 {
		q.TapRd = QdTrd
	}
	if q.VirSnd == 0 {
		q.VirSnd = QdVsd
	}
	if q.VirWrt == 0 {
		q.VirWrt = QdVWr
	}
	libol.Debug("Queue.Default %v", q)
}
