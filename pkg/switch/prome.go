package cswitch

import (
	"os"
	"strings"
	"time"

	"github.com/luscis/openlan/pkg/api"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	metrics = prometheus.NewRegistry()
	// Usage
	cpuTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_cpu_total",
		Help: "Current cpu cores",
	}, []string{"node"})
	cpuUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_cpu_usage",
		Help: "Current cpu usage in 100 percent",
	}, []string{"node"})
	memoryTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_memory_total_bytes",
		Help: "Current memory total in bytes",
	}, []string{"node"})
	memoryUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_memory_used_bytes",
		Help: "Current memory used in bytes",
	}, []string{"node"})
	diskTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_disk_total_bytes",
		Help: "Current disk total in bytes",
	}, []string{"node"})
	DiskUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_disk_used_bytes",
		Help: "Current disk used in bytes",
	}, []string{"node"})
	// Deices
	deviceSent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_device_transmit_bytes_total",
		Help: "Current device transmit bytes total",
	}, []string{"node", "name", "scope"})
	deviceRecv = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_device_received_bytes_total",
		Help: "Current device received bytes total",
	}, []string{"node", "name", "scope"})
	// Clients
	clientSent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_client_transmit_bytes_total",
		Help: "Current client transmit bytes total",
	}, []string{"node", "name", "scope", "address"})
	clientRecv = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_client_received_bytes_total",
		Help: "Current client received bytes total",
	}, []string{"node", "name", "scope", "address"})
)

func nodeName() string {
	if node, err := os.Hostname(); err == nil {
		return node
	}
	return "default"
}

func updateUsage() {
	usage := api.GetUsage()
	labels := prometheus.Labels{"node": nodeName()}
	cpuTotal.With(labels).Set(float64(usage.CPUTotal))
	cpuUsage.With(labels).Set(float64(usage.CPUUsage))
	memoryTotal.With(labels).Set(float64(usage.MemTotal))
	memoryUsed.With(labels).Set(float64(usage.MemUsed))
	diskTotal.With(labels).Set(float64(usage.DiskTotal))
	DiskUsed.With(labels).Set(float64(usage.DiskUsed))
}

func updateDevices() {
	devices := api.ListDevices()
	deviceSent.Reset()
	deviceRecv.Reset()
	for _, d := range devices {
		labels := prometheus.Labels{
			"node":  nodeName(),
			"name":  d.Name,
			"scope": d.Network,
		}
		deviceSent.With(labels).Set(float64(d.Send))
		deviceRecv.With(labels).Set(float64(d.Recv))
	}
}

func updateClients() {
	clients := api.ListClients()
	clientSent.Reset()
	clientRecv.Reset()
	for _, c := range clients {
		labels := prometheus.Labels{
			"node":    nodeName(),
			"name":    strings.SplitN(c.Name, "@", 2)[0],
			"scope":   c.Network,
			"address": c.Address,
		}
		clientSent.With(labels).Set(float64(c.TxBytes))
		clientRecv.With(labels).Set(float64(c.RxBytes))
	}
}

func recordMetrics() {
	libol.Go(func() {
		for {
			updateUsage()
			updateDevices()
			updateClients()
			time.Sleep(2 * time.Second)
		}
	})
}

func init() {
	recordMetrics()
	metrics.MustRegister(cpuTotal)
	metrics.MustRegister(cpuUsage)
	metrics.MustRegister(memoryTotal)
	metrics.MustRegister(memoryUsed)
	metrics.MustRegister(diskTotal)
	metrics.MustRegister(DiskUsed)
	metrics.MustRegister(deviceSent)
	metrics.MustRegister(deviceRecv)
	metrics.MustRegister(clientSent)
	metrics.MustRegister(clientRecv)
}
