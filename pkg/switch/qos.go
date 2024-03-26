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
	inSpeed  int64 // bits
	outSpeed int64 // bits
	name     string
	ip       string
	device   string
	chainIn  *cn.FireWallChain
	chainOut *cn.FireWallChain
	out      *libol.SubLogger
}

func (qr *QosRule) Name(dir string) string {
	return "Qos_" + qr.name + "_" + dir
}
func (qr *QosRule) InLimitPacket() string {
	//bytes / mtu
	return strconv.Itoa(int(qr.inSpeed / 1500))
}
func (qr *QosRule) OutLimitPacket() string {
	//bytes / mtu
	return strconv.Itoa(int(qr.outSpeed / 1500))
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
		Limit: qr.OutLimitStr(),
		//LimitBurst: qr.OutLimitPacket(),
		Comment: "Qos Limit",
		Jump:    "ACCEPT",
	}
}

func (qr *QosRule) buildChainOut(fire *cn.FireWallTable) {
	if qr.outSpeed > 0 {
		qr.chainOut = cn.NewFireWallChain(qr.Name("out"), cn.TMangle, "")
		qr.chainOut.AddRule(qr.OutLimitRule())
		qr.chainOut.AddRule(cn.IPRule{
			Comment: "Qos Default Drop",
			Jump:    "DROP",
		})
		qr.chainOut.Install()

		qr.buildChainOutJump(fire)
	}
}
func (qr *QosRule) buildChainOutJump(fire *cn.FireWallTable) {
	if qr.device != "" && qr.ip != "" {
		if err := fire.Mangle.Out.AddRuleX(cn.IPRule{
			Comment: "Qos Forward",
			Jump:    qr.Name("out"),
			Output:  qr.device,
			Dest:    qr.ip,
		}); err != nil {
			qr.out.Warn("Qos.Add Out Rule: %s", err)
		}
	}
}
func (qr *QosRule) clearChainOutJump(fire *cn.FireWallTable) {
	if err := fire.Mangle.Out.DelRuleX(cn.IPRule{
		Comment: "Qos Forward",
		Jump:    qr.Name("out"),
		Output:  qr.device,
		Dest:    qr.ip,
	}); err != nil {
		qr.out.Warn("Qos.Del Out Rule: %s", err)
	}
}
func (qr *QosRule) buildChainIn(fire *cn.FireWallTable) {
	if qr.inSpeed > 0 {
		qr.chainIn = cn.NewFireWallChain(qr.Name("in"), cn.TMangle, "")
		qr.chainIn.AddRule(qr.InLimitRule())
		qr.chainIn.AddRule(cn.IPRule{
			Comment: "Qos Default Drop",
			Jump:    "DROP",
		})
		qr.chainIn.Install()

		qr.buildChainInJump(fire)
	}
}
func (qr *QosRule) buildChainInJump(fire *cn.FireWallTable) {
	if qr.device != "" && qr.ip != "" {
		if err := fire.Mangle.In.AddRuleX(cn.IPRule{
			Comment: "Qos Forward",
			Jump:    qr.Name("in"),
			Input:   qr.device,
			Source:  qr.ip,
		}); err != nil {
			qr.out.Warn("Qos.Add In Rule: %s", err)
		}
	}
}
func (qr *QosRule) clearChainInJump(fire *cn.FireWallTable) {
	if err := fire.Mangle.In.DelRuleX(cn.IPRule{
		Comment: "Qos Forward",
		Jump:    qr.Name("in"),
		Input:   qr.device,
		Source:  qr.ip,
	}); err != nil {
		qr.out.Warn("Qos.Del In Rule: %s", err)
	}
}
func (qr *QosRule) Initialize(fire *cn.FireWallTable) {
	qr.buildChainIn(fire)
	qr.buildChainOut(fire)
}

func (qr *QosRule) reBuild(fire *cn.FireWallTable) {
	qr.Clear(fire)
	qr.Initialize(fire)
}
func (qr *QosRule) ClearChainIn(fire *cn.FireWallTable) {
	if qr.chainIn != nil {
		qr.clearChainInJump(fire)
		qr.chainIn.Cancel()
		qr.chainIn = nil
	}
}
func (qr *QosRule) ClearChainOut(fire *cn.FireWallTable) {
	if qr.chainOut != nil {
		qr.clearChainOutJump(fire)
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
	if qr.device != device || qr.ip != ip {
		changeDeviceOrIp = true
		qr.device = device
		qr.ip = ip
	}

	if changeDeviceOrIp {
		qr.clearChainInJump(fire)
		qr.clearChainOutJump(fire)
		qr.buildChainOutJump(fire)
		qr.buildChainInJump(fire)
	}

	if qr.inSpeed != inSpeed {
		qr.inSpeed = inSpeed
		qr.ClearChainIn(fire)
		qr.buildChainIn(fire)
	}

	if qr.outSpeed != outSpeed {
		qr.outSpeed = outSpeed
		qr.ClearChainOut(fire)
		qr.buildChainOut(fire)
	}

}

type QosCtrl struct {
	rules map[string]*QosRule
	fire  *cn.FireWallTable
	out   *libol.SubLogger
	lock  sync.Mutex
}

func NewQosCtrl(fire *cn.FireWallTable) *QosCtrl {
	return &QosCtrl{
		fire:  fire,
		rules: make(map[string]*QosRule, 1024),
		out:   libol.NewSubLogger("Qos"),
	}
}

func (q *QosCtrl) Initialize(qos *config.Qos) {
	//q.Start()
	if qos != nil && qos.Config != nil {
		for name, limit := range qos.Config {
			rule := &QosRule{
				name:     name,
				inSpeed:  limit.InSpeed,
				outSpeed: limit.OutSpeed,
				device:   "",
				ip:       "",
				out:      libol.NewSubLogger("Qos_" + name),
			}
			q.rules[name] = rule
		}
	}
}

func (q *QosCtrl) Start() {
	q.out.Info("Qos.Start")
	if len(q.rules) != 0 {
		for _, rule := range q.rules {
			rule.Initialize(q.fire)
		}
	}
	libol.Go(q.Update)
}

func (q *QosCtrl) Stop() {
	q.out.Info("Qos.Stop")
	if len(q.rules) != 0 {
		for _, rule := range q.rules {
			rule.Clear(q.fire)
		}
	}
}

func (q *QosCtrl) DelUserRule(name string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if rule, ok := q.rules[name]; ok {
		rule.Clear(q.fire)
		delete(q.rules, name)
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

	if rule, ok := q.rules[name]; ok {

		rule.Update(q.fire, inSpeed, outSpeed, device, ip)
	} else {

		rule = &QosRule{
			name:     name,
			inSpeed:  inSpeed,
			outSpeed: outSpeed,
			device:   device,
			ip:       ip,
			out:      libol.NewSubLogger("Qos_" + name),
		}
		rule.Initialize(q.fire)

		q.rules[name] = rule
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
	for _, rule := range q.rules {
		var existClient *schema.VPNClient
		for _, client := range clients {
			if client.Name == rule.name {
				existClient = &client
				break
			}
		}
		if existClient != nil {
			rule.Update(q.fire, rule.inSpeed, rule.outSpeed, existClient.Device, existClient.Address)
		} else {
			if rule.device != "" || rule.ip != "" {
				rule.clearChainInJump(q.fire)
				rule.clearChainOutJump(q.fire)
				rule.device = ""
				rule.ip = ""
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

	for _, rule := range q.rules {
		obj := schema.Qos{
			Name:     rule.name,
			InSpeed:  rule.inSpeed,
			OutSpeed: rule.outSpeed,
			Device:   rule.device,
			Ip:       rule.ip,
		}
		call(obj)
	}
}
