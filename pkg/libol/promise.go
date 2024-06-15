package libol

import "time"

type Promise struct {
	Count  int
	MaxTry int
	First  time.Duration // the delay time.
	MinInt time.Duration // the normal time.
	MaxInt time.Duration // the max delay time.
}

func NewPromise() *Promise {
	return &Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
		MaxTry: 10,
	}
}

func NewPromiseAlways() *Promise {
	return &Promise{
		First:  time.Second * 2,
		MaxInt: time.Minute,
		MinInt: time.Second * 10,
	}
}

func (p *Promise) Done(call func() error) {
	for {
		p.Count++
		if p.MaxTry > 0 && p.Count > p.MaxTry {
			return
		}
		if err := call(); err == nil {
			return
		}
		time.Sleep(p.First)
		if p.First < p.MaxInt {
			p.First += p.MinInt
		}
	}
}

func (p *Promise) Go(call func() error) {
	Go(func() {
		p.Done(call)
	})
}
