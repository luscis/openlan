package cswitch

import (
	"time"

	"github.com/luscis/openlan/pkg/libol"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metrics     = prometheus.NewRegistry()
	helloMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hello_ops_total",
		Help: "The total hello number",
	})
)

func recordMetrics() {
	libol.Go(func() {
		for {
			helloMetric.Inc()
			time.Sleep(2 * time.Second)
		}
	})
}

func init() {
	recordMetrics()
	metrics.MustRegister(helloMetric)
}
