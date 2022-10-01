package access

import (
	"github.com/luscis/openlan/pkg/access/http"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/network"
	"runtime"
)

type Pointer interface {
	Addr() string
	IfName() string
	IfAddr() string
	Client() libol.SocketClient
	Device() network.Taper
	Status() libol.SocketStatus
	UpTime() int64
	UUID() string
	Protocol() string
	User() string
	Record() map[string]int64
	Tenant() string
	Alias() string
	Config() *config.Point
	Network() *models.Network
}

type MixPoint struct {
	uuid   string
	worker *Worker
	config *config.Point
	out    *libol.SubLogger
	http   *http.Http
}

func NewMixPoint(config *config.Point) MixPoint {
	return MixPoint{
		worker: NewWorker(config),
		config: config,
		out:    libol.NewSubLogger(config.Id()),
	}
}

func (p *MixPoint) Initialize() {
	libol.Info("MixPoint.Initialize")
	p.worker.SetUUID(p.UUID())
	p.worker.Initialize()
	if p.config.Http != nil {
		p.http = http.NewHttp(p)
	}
}

func (p *MixPoint) Start() {
	p.out.Info("MixPoint.Start %s", runtime.GOOS)
	if p.config.PProf != "" {
		f := libol.PProf{Listen: p.config.PProf}
		f.Start()
	}
	p.worker.Start()
}

func (p *MixPoint) Stop() {
	defer libol.Catch("MixPoint.Stop")
	if p.http != nil {
		p.http.Shutdown()
	}
	p.worker.Stop()
}

func (p *MixPoint) UUID() string {
	if p.uuid == "" {
		p.uuid = libol.GenString(13)
	}
	return p.uuid
}

func (p *MixPoint) Status() libol.SocketStatus {
	client := p.Client()
	if client == nil {
		return 0
	}
	return client.Status()
}

func (p *MixPoint) Addr() string {
	return p.config.Connection
}

func (p *MixPoint) IfName() string {
	device := p.Device()
	if device == nil {
		return ""
	}
	return device.Name()
}

func (p *MixPoint) Client() libol.SocketClient {
	if p.worker.conWorker == nil {
		return nil
	}
	return p.worker.conWorker.client
}

func (p *MixPoint) Device() network.Taper {
	if p.worker.tapWorker == nil {
		return nil
	}
	return p.worker.tapWorker.device
}

func (p *MixPoint) UpTime() int64 {
	return p.worker.UpTime()
}

func (p *MixPoint) IfAddr() string {
	return p.worker.ifAddr
}

func (p *MixPoint) Tenant() string {
	return p.config.Network
}

func (p *MixPoint) User() string {
	return p.config.Username
}

func (p *MixPoint) Alias() string {
	return p.config.Alias
}

func (p *MixPoint) Record() map[string]int64 {
	rt := p.worker.conWorker.record
	// TODO padding data from tapWorker
	return rt.Data()
}

func (p *MixPoint) Config() *config.Point {
	return p.config
}

func (p *MixPoint) Network() *models.Network {
	return p.worker.network
}

func (p *MixPoint) Protocol() string {
	return p.config.Protocol
}
