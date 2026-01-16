package cswitch

import (
	"time"

	"github.com/luscis/openlan/pkg/api"
	"github.com/luscis/openlan/pkg/cache"
	co "github.com/luscis/openlan/pkg/config"
	cn "github.com/luscis/openlan/pkg/network"
)

func PeerName(name, prefix string) (string, string) {
	return name + prefix + "i", name + prefix + "o"
}

type OpenLANWorker struct {
	*WorkerImpl
	alias     string
	newTime   int64
	startTime int64
	links     *Links
}

func NewOpenLANWorker(c *co.Network) *OpenLANWorker {
	return &OpenLANWorker{
		WorkerImpl: NewWorkerApi(c),
		alias:      c.Alias,
		newTime:    time.Now().Unix(),
		startTime:  0,
		links:      NewLinks(),
	}
}

func (w *OpenLANWorker) Initialize() {
	brCfg := w.cfg.Bridge
	name := w.cfg.Name

	w.br = cn.NewBridger(brCfg.Provider, brCfg.Name, brCfg.IPMtu)

	for _, ht := range w.cfg.Hosts {
		lease := cache.Network.AddLease(ht.Hostname, ht.Address, name)
		if lease != nil {
			lease.Type = "static"
			lease.Network = name
		}
	}

	w.WorkerImpl.Initialize()
}

func (w *OpenLANWorker) UpBridge(cfg *co.Bridge) {
	master := w.br
	// new it and configure address
	master.Open(cfg.Address)
	// configure stp
	if cfg.Stp == "enable" {
		if err := master.Stp(true); err != nil {
			w.out.Warn("OpenLANWorker.UpBridge: Stp %s", err)
		}
	} else {
		_ = master.Stp(false)
	}
	// configure forward delay
	if err := master.Delay(cfg.Delay); err != nil {
		w.out.Warn("OpenLANWorker.UpBridge: Delay %s", err)
	}
	if err := master.CallIptables(0); err != nil {
		w.out.Warn("OpenLANWorker.Start: CallIptables %s", err)
	}
}

func (w *OpenLANWorker) Start(v api.SwitchApi) {
	w.uuid = v.UUID()
	w.startTime = time.Now().Unix()
	w.out.Info("OpenLANWorker.Start")
	w.UpBridge(w.cfg.Bridge)
	w.WorkerImpl.Start(v)
}

func (w *OpenLANWorker) downBridge() {
	if err := w.br.Close(); err != nil {
		w.out.Warn("OpenLANWorker.downBridge: %s", err)
	}
}

func (w *OpenLANWorker) Stop(kill bool) {
	w.out.Info("OpenLANWorker.Close")
	w.WorkerImpl.Stop(kill)
	w.startTime = 0
	if kill {
		w.downBridge()
	}
}

func (w *OpenLANWorker) UpTime() int64 {
	if w.startTime != 0 {
		return time.Now().Unix() - w.startTime
	}
	return 0
}
