package access

import (
	"runtime"

	"github.com/luscis/openlan/pkg/access/http"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/network"
)

type Acceser interface {
	Addr() string
	IfName() string
	IfAddr() string
	Status() libol.SocketStatus
	UpTime() int64
	UUID() string
	Protocol() string
	User() string
	Record() map[string]int64
	Statistics() map[string]int64
	Tenant() string
	Alias() string
	Config() *config.Access
	Network() *models.Network
}

type MixAccess struct {
	uuid    string
	worker  *Worker
	config  *config.Access
	out     *libol.SubLogger
	http    *http.Http
	addr    string
	gateway string
	brName  string
}

func NewMixAccess(config *config.Access) MixAccess {
	return MixAccess{
		worker: NewWorker(config),
		config: config,
		out:    libol.NewSubLogger(config.Id()),
		brName: config.Interface.Bridge,
	}
}

func (p *MixAccess) Initialize() {
	libol.Info("MixAccess.Initialize")
	p.worker.SetUUID(p.UUID())
	p.worker.Initialize()
	if p.config.Http != nil {
		p.http = http.NewHttp(p)
	}
}

func (p *MixAccess) Start() {
	p.Run0()
	p.out.Info("MixAccess.Start %s", runtime.GOOS)
	if p.config.PProf != "" {
		f := libol.PProf{Listen: p.config.PProf}
		f.Start()
	}
	p.worker.Start()
}

func (p *MixAccess) Stop() {
	defer libol.Catch("MixAccess.Stop")
	if p.http != nil {
		p.http.Shutdown()
	}
	p.worker.Stop()
}

func (p *MixAccess) UUID() string {
	if p.uuid == "" {
		p.uuid = libol.GenString(13)
	}
	return p.uuid
}

func (p *MixAccess) Status() libol.SocketStatus {
	client := p.client()
	if client == nil {
		return 0
	}
	return client.Status()
}

func (p *MixAccess) Addr() string {
	return p.config.Connection
}

func (p *MixAccess) IfName() string {
	device := p.device()
	if device == nil {
		return ""
	}
	return device.Name()
}

func (p *MixAccess) client() libol.SocketClient {
	conn := p.worker.conWorker
	if conn == nil {
		return nil
	}
	return conn.client
}

func (p *MixAccess) device() network.Taper {
	tap := p.worker.tapWorker
	if tap == nil {
		return nil
	}
	return tap.device
}

func (p *MixAccess) UpTime() int64 {
	return p.worker.UpTime()
}

func (p *MixAccess) IfAddr() string {
	return p.worker.ifAddr
}

func (p *MixAccess) Tenant() string {
	return p.config.Network
}

func (p *MixAccess) User() string {
	return p.config.Username
}

func (p *MixAccess) Alias() string {
	return p.config.Alias
}

func (p *MixAccess) Record() map[string]int64 {
	rt := p.worker.conWorker.record
	// TODO padding data from tapWorker
	return rt.Data()
}

func (p *MixAccess) Config() *config.Access {
	return p.config
}

func (p *MixAccess) Protocol() string {
	return p.config.Protocol
}

func (p *MixAccess) Statistics() map[string]int64 {
	client := p.client()
	if client == nil {
		return nil
	}
	return client.Statistics()
}

func (p *MixAccess) Run1() {
	bin := p.config.Run1
	if bin == "" {
		return
	}
	if out, err := libol.Exec(bin); err != nil {
		p.out.Warn("MixAccess.Run1: %s, %s", err, out)
	} else {
		p.out.Info("MixAccess.Run1: %s", out)
	}
}

func (p *MixAccess) Run0() {
	bin := p.config.Run0
	if bin == "" {
		return
	}
	if out, err := libol.Exec(bin); err != nil {
		p.out.Warn("MixAccess.Run0: %s, %s", err, out)
	} else {
		p.out.Info("MixAccess.Run0: %s", out)
	}
}
