package cswitch

import (
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
	"strconv"
	"sync"
	"time"
)

//125000 ~ 1Mb/s

type QosRule struct {
	InSpeed  int64 // bits
	OutSpeed int64 // bits
	Name     string
	Ip       string
	Device   string
	chainIn  *cn.FireWallChain
	chainOut *cn.FireWallChain
	out      *libol.SubLogger
}

func (qr *QosRule) RuleName(dir string) string {
	return "Qos_" + qr.Name + "_" + dir
}
func (qr *QosRule) InLimitPacket() string {
	//bytes / mtu
	return strconv.Itoa(int(qr.InSpeed / 1500))
}
func (qr *QosRule) OutLimitPacket() string {
	//bytes / mtu
	return strconv.Itoa(int(qr.OutSpeed / 1500))
}

func (qr *QosRule) InLimitStr() string {
	//bytes / mtu
	return qr.InLimitPacket() + "/s"
}
func (qr *QosRule) OutLimitStr() string {
	//bytes / mtu
	return qr.OutLimitPacket() + "/s"
}

func (qr *QosRule) InLimitRule() cn.IPRule {
	return cn.IPRule{
		Limit: qr.InLimitStr(),
		//LimitBurst: qr.InLimitPacket(),
		Comment: "Qos Limit",
		Jump:    "ACCEPT",
	}
}

func (qr *QosRule) OutLimitRule() cn.IPRule {
	return cn.IPRule{
		Limit:   qr.OutLimitStr(),
		Comment: "Qos Limit",
		Jump:    "ACCEPT",
	}
}

func (qr *QosRule) BuildChainOut(fire *cn.FireWallTable) {
	if qr.OutSpeed > 0 {
		qr.chainOut = cn.NewFireWallChain(qr.RuleName("out"), cn.TMangle, "")
		qr.chainOut.AddRule(qr.OutLimitRule())
		qr.chainOut.AddRule(cn.IPRule{
			Comment: "Qos Default Drop",
			Jump:    "DROP",
		})
		qr.chainOut.Install()

		qr.BuildChainOutJump(fire)
	}
}

func (qr *QosRule) BuildChainOutJump(fire *cn.FireWallTable) {
	if qr.Device != "" && qr.Ip != "" {
		if err := fire.Mangle.Out.AddRuleX(cn.IPRule{
			Comment: "Qos Jump",
			Jump:    qr.RuleName("out"),
			Output:  qr.Device,
			Dest:    qr.Ip,
		}); err != nil {
			qr.out.Warn("Qos.Add Out Rule: %s", err)
		}
	}
}

func (qr *QosRule) ClearChainOutJump(fire *cn.FireWallTable) {
	if err := fire.Mangle.Out.DelRuleX(cn.IPRule{
		Comment: "Qos Jump",
		Jump:    qr.RuleName("out"),
		Output:  qr.Device,
		Dest:    qr.Ip,
	}); err != nil {
		qr.out.Warn("Qos.Del Out Rule: %s", err)
	}
}

func (qr *QosRule) BuildChainIn(fire *cn.FireWallTable) {
	if qr.InSpeed > 0 {
		qr.chainIn = cn.NewFireWallChain(qr.RuleName("in"), cn.TMangle, "")
		qr.chainIn.AddRule(qr.InLimitRule())
		qr.chainIn.AddRule(cn.IPRule{
			Comment: "Qos Default Drop",
			Jump:    "DROP",
		})
		qr.chainIn.Install()

		qr.BuildChainInJump(fire)
	}
}

func (qr *QosRule) BuildChainInJump(fire *cn.FireWallTable) {
	if qr.Device != "" && qr.Ip != "" {
		if err := fire.Mangle.In.AddRuleX(cn.IPRule{
			Comment: "Qos Forward",
			Jump:    qr.RuleName("in"),
			Input:   qr.Device,
			Source:  qr.Ip,
		}); err != nil {
			qr.out.Warn("Qos.Add In Rule: %s", err)
		}
	}
}

func (qr *QosRule) ClearChainInJump(fire *cn.FireWallTable) {
	if err := fire.Mangle.In.DelRuleX(cn.IPRule{
		Comment: "Qos Forward",
		Jump:    qr.RuleName("in"),
		Input:   qr.Device,
		Source:  qr.Ip,
	}); err != nil {
		qr.out.Warn("Qos.Del In Rule: %s", err)
	}
}

func (qr *QosRule) Start(fire *cn.FireWallTable) {
	qr.BuildChainIn(fire)
	qr.BuildChainOut(fire)
}

func (qr *QosRule) ReBuild(fire *cn.FireWallTable) {
	qr.Clear(fire)
	qr.Start(fire)
}
func (qr *QosRule) ClearChainIn(fire *cn.FireWallTable) {
	if qr.chainIn != nil {
		qr.ClearChainInJump(fire)
		qr.chainIn.Cancel()
		qr.chainIn = nil
	}
}
func (qr *QosRule) ClearChainOut(fire *cn.FireWallTable) {
	if qr.chainOut != nil {
		qr.ClearChainOutJump(fire)
		qr.chainOut.Cancel()
		qr.chainOut = nil
	}
}
func (qr *QosRule) Clear(fire *cn.FireWallTable) {
	qr.ClearChainIn(fire)
	qr.ClearChainOut(fire)
}

func (qr *QosRule) Update(fire *cn.FireWallTable, inSpeed int64, outSpeed int64, device string, ip string) {

	var changeDeviceOrIp = false
	if qr.Device != device || qr.Ip != ip {
		changeDeviceOrIp = true
		qr.Device = device
		qr.Ip = ip
	}

	if changeDeviceOrIp {
		qr.ClearChainInJump(fire)
		qr.ClearChainOutJump(fire)
		qr.BuildChainOutJump(fire)
		qr.BuildChainInJump(fire)
	}

	if qr.InSpeed != inSpeed {
		qr.InSpeed = inSpeed
		qr.ClearChainIn(fire)
		qr.BuildChainIn(fire)
	}

	if qr.OutSpeed != outSpeed {
		qr.OutSpeed = outSpeed
		qr.ClearChainOut(fire)
		qr.BuildChainOut(fire)
	}

}

type QosCtrl struct {
	Name  string
	Rules map[string]*QosRule
	fire  *cn.FireWallTable
	out   *libol.SubLogger
	lock  sync.Mutex
}

func NewQosCtrl(fire *cn.FireWallTable, name string) *QosCtrl {
	return &QosCtrl{
		Name:  name,
		Rules: make(map[string]*QosRule, 1024),
		fire:  fire,
		out:   libol.NewSubLogger("Qos"),
	}
}

func (q *QosCtrl) Initialize() {
	//q.Start()
	qosCfg := config.GetQos(q.Name)

	if qosCfg != nil && len(qosCfg.Config) > 0 {
		for name, limit := range qosCfg.Config {
			qr := &QosRule{
				Name:     name,
				InSpeed:  limit.InSpeed,
				OutSpeed: limit.OutSpeed,
				Device:   "",
				Ip:       "",
				out:      libol.NewSubLogger("Qos_" + name),
			}
			q.Rules[name] = qr
		}

	}
}

func (q *QosCtrl) Start() {
	q.out.Info("Qos.Start")
	if len(q.Rules) > 0 {
		for _, rule := range q.Rules {
			rule.Start(q.fire)
		}
	}
	libol.Go(q.Update)
}

func (q *QosCtrl) Stop() {
	q.out.Info("Qos.Stop")
	if len(q.Rules) != 0 {
		for _, rule := range q.Rules {
			rule.Clear(q.fire)
		}
	}
}

func (q *QosCtrl) DelUserRule(name string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if rule, ok := q.Rules[name]; ok {
		rule.Clear(q.fire)
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

func (q *QosCtrl) AddOrUserRule(name string, inSpeed int64, outSpeed int64) {
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

		rule.Update(q.fire, inSpeed, outSpeed, device, ip)
	} else {

		rule = &QosRule{
			Name:     name,
			InSpeed:  inSpeed,
			OutSpeed: outSpeed,
			Device:   device,
			Ip:       ip,
			out:      libol.NewSubLogger("Qos_" + name),
		}
		rule.Start(q.fire)

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
			rule.Update(q.fire, rule.InSpeed, rule.OutSpeed, existClient.Device, existClient.Address)
		} else {
			if rule.Device != "" || rule.Ip != "" {
				rule.ClearChainInJump(q.fire)
				rule.ClearChainOutJump(q.fire)
				rule.Device = ""
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
			InSpeed:  rule.InSpeed,
			OutSpeed: rule.OutSpeed,
		}
		cfg.Config[rule.Name] = ql
	}
	cfg.Save()
}

func (q *QosCtrl) AddUserQos(name string, inSpeed int64, outSpeed int64) error {

	q.AddOrUserRule(name, inSpeed, outSpeed)

	return nil
}
func (q *QosCtrl) UpdateUserQos(name string, inSpeed int64, outSpeed int64) error {

	q.AddOrUserRule(name, inSpeed, outSpeed)

	return nil
}
func (q *QosCtrl) DelUserQos(name string) error {

	q.DelUserRule(name)
	return nil
}

func (q *QosCtrl) ListQosUsers(call func(obj schema.Qos)) {

	for _, rule := range q.Rules {
		obj := schema.Qos{
			Name:     rule.Name,
			InSpeed:  rule.InSpeed,
			OutSpeed: rule.OutSpeed,
			Device:   rule.Device,
			Ip:       rule.Ip,
		}
		call(obj)
	}
}
