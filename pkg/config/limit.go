package config

import "github.com/luscis/openlan/pkg/libol"

type Queue struct {
	SockWr int `json:"sockWr" yaml:"sockWr"` // per frames about 1572(1514+4+20+20+14)bytes
	SockRd int `json:"sockRd" yaml:"sockRd"` // per frames
	TapWr  int `json:"tapWr" yaml:"tapWr"`   // per frames about 1572((1514+4+20+20+14))bytes
	TapRd  int `json:"tapRd" yaml:"tapRd"`   // per frames
	VirSnd int `json:"virSnd" yaml:"virSnd"`
	VirWrt int `json:"virWrt" yaml:"virWrt"`
}

var (
	QdSwr = 32 * 4
	QdSrd = 32 * 4
	QdTwr = 32 * 2
	QdTrd = 2
	QdVsd = 32 * 8
	QdVWr = 32 * 4
)

func (q *Queue) Correct() {
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
	libol.Info("Queue.Correct %v", q)
}
