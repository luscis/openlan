package cswitch

import (
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
	"strconv"
	"strings"
	"sync"
	"time"
)

//125000 ~ 1Mb/s

type QosUser struct {
	QosChainName string
	InSpeed      float64 // Mbit
	Name         string
	Ip           string
	Device       string
	qosChainIn   *cn.FireWallChain
	out          *libol.SubLogger
}

func (qr *QosUser) RuleName(dir string) string {
	nameParts := strings.Split(qr.Name, "@")
	return "Qos_" + qr.QosChainName + "-" + dir + "-" + nameParts[0]
}

func (qr *QosUser) InLimitPacket() string {
	//Mbit * 125000 / mtu
	return strconv.Itoa(int((qr.InSpeed * 125000) / 1300))
}

func (qr *QosUser) InLimitStr() string {
	return qr.InLimitPacket() + "/s"
}

func (qr *QosUser) InLimitRule() cn.IPRule {
	return cn.IPRule{
		Limit:      qr.InLimitStr(),
		LimitBurst: "100",
		Comment:    "Qos Limit In " + qr.Name,
		Jump:       "ACCEPT",
	}
}

func (qr *QosUser) BuildChainIn(chain *cn.FireWallChain) {
	if qr.InSpeed > 0 {
		qr.qosChainIn = cn.NewFireWallChain(qr.RuleName("in"), cn.TMangle, "")
		qr.qosChainIn.AddRule(qr.InLimitRule())
		qr.qosChainIn.AddRule(cn.IPRule{
			Comment: "Qos Default Drop",
			Jump:    "DROP",
		})
		qr.qosChainIn.Install()

		qr.BuildChainInJump(chain)
	}
}

func (qr *QosUser) BuildChainInJump(chain *cn.FireWallChain) {
	if qr.Ip != "" {
		if err := chain.AddRuleX(cn.IPRule{
			Comment: "Qos Jump",
			Jump:    qr.RuleName("in"),
			Source:  qr.Ip,
		}); err != nil {
			qr.out.Warn("Qos.Add In Rule: %s", err)
		}
	}
}

func (qr *QosUser) ClearChainInJump(chain *cn.FireWallChain) {
	if err := chain.DelRuleX(cn.IPRule{
		Comment: "Qos Jump",
		Jump:    qr.RuleName("in"),
		Source:  qr.Ip,
	}); err != nil {
		qr.out.Warn("Qos.Del In Rule: %s", err)
	}
}

func (qr *QosUser) Start(chainIn *cn.FireWallChain) {
	qr.BuildChainIn(chainIn)
}

func (qr *QosUser) ReBuild(chainIn *cn.FireWallChain) {
	qr.Clear(chainIn)
	qr.Start(chainIn)
}

func (qr *QosUser) ClearChainIn(chain *cn.FireWallChain) {
	if qr.qosChainIn != nil {
		qr.ClearChainInJump(chain)
		qr.qosChainIn.Cancel()
		qr.qosChainIn = nil
	}
}

func (qr *QosUser) Clear(chainIn *cn.FireWallChain) {
	qr.ClearChainIn(chainIn)
}

func (qr *QosUser) Update(chainIn *cn.FireWallChain, inSpeed float64, device string, ip string) {

	changed := false
	qr.Device = device
	if qr.Ip != ip {
		changed = true
		qr.Ip = ip
	}

	if changed {
		qr.ClearChainInJump(chainIn)
		qr.BuildChainInJump(chainIn)
	}

	if qr.InSpeed != inSpeed {
		qr.InSpeed = inSpeed
		qr.ClearChainIn(chainIn)
		qr.BuildChainIn(chainIn)
	}

}

type QosCtrl struct {
	Name    string
	Rules   map[string]*QosUser
	chainIn *cn.FireWallChain
	out     *libol.SubLogger
	lock    sync.Mutex
}

func NewQosCtrl(name string) *QosCtrl {
	return &QosCtrl{
		Name:  name,
		Rules: make(map[string]*QosUser, 1024),
		out:   libol.NewSubLogger("Qos"),
	}
}

func (q *QosCtrl) ChainIn() string {
	return "Qos_" + q.Name + "-in"
}

func (q *QosCtrl) Initialize() {
	//q.Start()
	q.chainIn = cn.NewFireWallChain(q.ChainIn(), cn.TMangle, "")

	qosCfg := config.GetQos(q.Name)

	if qosCfg != nil && len(qosCfg.Config) > 0 {
		for name, limit := range qosCfg.Config {
			qr := &QosUser{
				QosChainName: q.Name,
				Name:         name,
				InSpeed:      limit.InSpeed,
				Ip:           "",
				out:          libol.NewSubLogger("Qos_" + name),
			}
			q.Rules[name] = qr
		}

	}
}

func (q *QosCtrl) Start() {
	q.out.Info("Qos.Start")
	q.chainIn.Install()

	if len(q.Rules) > 0 {
		for _, rule := range q.Rules {
			rule.Start(q.chainIn)
		}
	}

	libol.Go(q.Update)
}

func (q *QosCtrl) Stop() {
	q.out.Info("Qos.Stop")
	if len(q.Rules) != 0 {
		for _, rule := range q.Rules {
			rule.Clear(q.chainIn)
		}
	}

	q.chainIn.Cancel()
}

func (q *QosCtrl) DelUserRule(name string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if rule, ok := q.Rules[name]; ok {
		rule.Clear(q.chainIn)
		delete(q.Rules, name)
	}
}

func (q *QosCtrl) FindClient(name string) *schema.VPNClient {
	for n := range cache.Network.List() {
		if n == nil {
			break
		}
		for client := range cache.VPNClient.List(n.Name) {
			if client == nil {
				break
			}
			if client.Name == name {
				return client
			}
		}
	}

	return nil
}

func (q *QosCtrl) AddOrUpdateQosUser(name string, inSpeed float64) {
	q.lock.Lock()
	defer q.lock.Unlock()
	client := q.FindClient(name)
	device := ""
	var ip = ""
	if client != nil {
		device = client.Device
		ip = client.Address
	}

	if rule, ok := q.Rules[name]; ok {

		rule.Update(q.chainIn, inSpeed, device, ip)
	} else {

		rule = &QosUser{
			QosChainName: q.Name,
			Name:         name,
			InSpeed:      inSpeed,
			Ip:           ip,
			out:          libol.NewSubLogger("Qos_" + name),
		}
		rule.Start(q.chainIn)

		q.Rules[name] = rule
	}
}

func (q *QosCtrl) ClientUpdate() {
	clients := make([]schema.VPNClient, 0, 1024)
	for n := range cache.Network.List() {
		if n == nil {
			break
		}
		for client := range cache.VPNClient.List(n.Name) {
			if client == nil {
				break
			}
			clients = append(clients, *client)
		}
	}
	for _, rule := range q.Rules {
		var existClient *schema.VPNClient
		for _, client := range clients {
			if client.Name == rule.Name {
				existClient = &client
				break
			}
		}
		if existClient != nil {
			rule.Update(q.chainIn, rule.InSpeed, existClient.Device, existClient.Address)
		} else {
			if rule.Ip != "" {
				rule.ClearChainInJump(q.chainIn)
				rule.Ip = ""
			}
		}
	}

}

func (q *QosCtrl) Update() {

	for {
		q.ClientUpdate()

		time.Sleep(time.Second * 5)
	}

}

func (q *QosCtrl) Save() {
	cfg := config.GetQos(q.Name)
	cfg.Config = make(map[string]*config.QosLimit, 1024)
	for _, rule := range q.Rules {
		ql := &config.QosLimit{
			InSpeed: rule.InSpeed,
		}
		cfg.Config[rule.Name] = ql
	}
	cfg.Save()
}

func (q *QosCtrl) AddQosUser(name string, inSpeed float64) error {

	q.AddOrUpdateQosUser(name, inSpeed)

	return nil
}
func (q *QosCtrl) UpdateQosUser(name string, inSpeed float64) error {

	q.AddOrUpdateQosUser(name, inSpeed)

	return nil
}
func (q *QosCtrl) DelQosUser(name string) error {

	q.DelUserRule(name)
	return nil
}

func (q *QosCtrl) ListQosUsers(call func(obj schema.Qos)) {

	for _, rule := range q.Rules {
		obj := schema.Qos{
			Name:    rule.Name,
			Device:  rule.Device,
			InSpeed: rule.InSpeed,
			Ip:      rule.Ip,
		}
		call(obj)
	}
}
