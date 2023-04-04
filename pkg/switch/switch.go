package _switch

import (
	"encoding/json"
	"github.com/luscis/openlan/pkg/app"
	"github.com/luscis/openlan/pkg/cache"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/network"
	"net"
	"strings"
	"sync"
	"time"
)

func GetSocketServer(s *co.Switch) libol.SocketServer {
	crypt := s.Crypt
	block := libol.NewBlockCrypt(crypt.Algo, crypt.Secret)
	switch s.Protocol {
	case "kcp":
		c := libol.NewKcpConfig()
		c.Block = block
		c.Timeout = time.Duration(s.Timeout) * time.Second
		return libol.NewKcpServer(s.Listen, c)
	case "tcp":
		c := &libol.TcpConfig{
			Block:   block,
			Timeout: time.Duration(s.Timeout) * time.Second,
			RdQus:   s.Queue.SockRd,
			WrQus:   s.Queue.SockWr,
		}
		return libol.NewTcpServer(s.Listen, c)
	case "udp":
		c := &libol.UdpConfig{
			Block:   block,
			Timeout: time.Duration(s.Timeout) * time.Second,
		}
		return libol.NewUdpServer(s.Listen, c)
	case "ws":
		c := &libol.WebConfig{
			Block:   block,
			Timeout: time.Duration(s.Timeout) * time.Second,
			RdQus:   s.Queue.SockRd,
			WrQus:   s.Queue.SockWr,
		}
		return libol.NewWebServer(s.Listen, c)
	case "wss":
		c := &libol.WebConfig{
			Block:   block,
			Timeout: time.Duration(s.Timeout) * time.Second,
			RdQus:   s.Queue.SockRd,
			WrQus:   s.Queue.SockWr,
		}
		if s.Cert != nil {
			c.Cert = &libol.CertConfig{
				Crt: s.Cert.CrtFile,
				Key: s.Cert.KeyFile,
			}
		}
		return libol.NewWebServer(s.Listen, c)
	default:
		c := &libol.TcpConfig{
			Block:   block,
			Timeout: time.Duration(s.Timeout) * time.Second,
			RdQus:   s.Queue.SockRd,
			WrQus:   s.Queue.SockWr,
		}
		if s.Cert != nil {
			c.Tls = s.Cert.GetTlsCfg()
		}
		return libol.NewTcpServer(s.Listen, c)
	}
}

type Apps struct {
	Auth     *app.Access
	Request  *app.Request
	Neighbor *app.Neighbors
	OnLines  *app.Online
}

type Hook func(client libol.SocketClient, frame *libol.FrameMessage) error

type Switch struct {
	lock     sync.Mutex
	cfg      *co.Switch
	apps     Apps
	firewall *network.FireWall
	hooks    []Hook
	http     *Http
	server   libol.SocketServer
	worker   map[string]Networker
	uuid     string
	newTime  int64
	out      *libol.SubLogger
	confd    *ConfD
	l2tp     *L2TP
}

func NewSwitch(c *co.Switch) *Switch {
	server := GetSocketServer(c)
	v := &Switch{
		cfg:      c,
		firewall: network.NewFireWall(c.FireWall),
		worker:   make(map[string]Networker, 32),
		server:   server,
		newTime:  time.Now().Unix(),
		hooks:    make([]Hook, 0, 64),
		out:      libol.NewSubLogger(c.Alias),
	}
	v.confd = NewConfd(v)
	return v
}

func (v *Switch) Protocol() string {
	if v.cfg == nil {
		return ""
	}
	return v.cfg.Protocol
}

func (v *Switch) enablePort(protocol, port string) {
	v.out.Info("Switch.enablePort %s %s", protocol, port)
	// allowed forward between source and prefix.
	v.firewall.AddRule(network.IpRule{
		Table:   network.TFilter,
		Chain:   network.OLCInput,
		Proto:   protocol,
		Match:   "multiport",
		DstPort: port,
	})
}

func (v *Switch) enableFwd(input, output, source, prefix string) {
	v.out.Debug("Switch.enableFwd %s:%s %s:%s", input, output, source, prefix)
	// allowed forward between source and prefix.
	v.firewall.AddRule(network.IpRule{
		Table:  network.TFilter,
		Chain:  network.OLCForward,
		Input:  input,
		Output: output,
		Source: source,
		Dest:   prefix,
	})
	if source != prefix {
		v.firewall.AddRule(network.IpRule{
			Table:  network.TFilter,
			Chain:  network.OLCForward,
			Output: input,
			Input:  output,
			Source: prefix,
			Dest:   source,
		})
	}
}

func (v *Switch) enableMasq(input, output, source, prefix string) {
	if source == prefix {
		return
	}
	// enable masquerade from source to prefix.
	if prefix == "" || prefix == "0.0.0.0/0" {
		v.firewall.AddRule(network.IpRule{
			Table:  network.TNat,
			Chain:  network.OLCPost,
			Source: source,
			NoDest: source,
			Jump:   network.CMasq,
		})
	} else {
		v.firewall.AddRule(network.IpRule{
			Table:  network.TNat,
			Chain:  network.OLCPost,
			Source: source,
			Dest:   prefix,
			Jump:   network.CMasq,
		})
	}
}

func (v *Switch) enableSnat(input, output, source, prefix string) {
	if source == prefix {
		return
	}
	// enable masquerade from source to prefix.
	v.firewall.AddRule(network.IpRule{
		Table:    network.TNat,
		Chain:    network.OLCPost,
		ToSource: source,
		Dest:     prefix,
		Jump:     network.CSnat,
	})
}

func (v *Switch) preWorkerVPN(w Networker, vCfg *co.OpenVPN) {
	if w == nil || vCfg == nil {
		return
	}
	cfg := w.Config()
	routes := vCfg.Routes
	routes = append(routes, vCfg.Subnet)
	if addr := w.Subnet(); addr != "" {
		libol.Info("Switch.preWorkerVPN %s subnet %s", cfg.Name, addr)
		routes = append(routes, addr)
	}
	for _, rt := range cfg.Routes {
		addr := rt.Prefix
		if addr == "0.0.0.0/0" {
			vCfg.Push = append(vCfg.Push, "redirect-gateway def1")
			continue
		}
		if _, inet, err := net.ParseCIDR(addr); err == nil {
			routes = append(routes, inet.String())
		}
	}
	vCfg.Routes = routes
}

func (v *Switch) preWorker(w Networker) {
	cfg := w.Config()
	if cfg.OpenVPN != nil {
		v.preWorkerVPN(w, cfg.OpenVPN)
	}
	br := cfg.Bridge
	if br.Mss > 0 {
		v.firewall.AddRule(network.IpRule{
			Table:   network.TMangle,
			Chain:   network.OLCPost,
			Output:  br.Name,
			Proto:   "tcp",
			Match:   "tcp",
			TcpFlag: []string{"SYN,RST", "SYN"},
			Jump:    "TCPMSS",
			SetMss:  br.Mss,
		})
	}
}

func (v *Switch) enableAcl(acl, input string) {
	if input == "" {
		return
	}
	if acl != "" {
		v.firewall.AddRule(network.IpRule{
			Table: network.TRaw,
			Chain: network.OLCPre,
			Input: input,
			Jump:  acl,
		})
	}
}

func (v *Switch) preNetVPN0(nCfg *co.Network, vCfg *co.OpenVPN) {
	if nCfg == nil || vCfg == nil {
		return
	}
	devName := vCfg.Device
	v.enableAcl(nCfg.Acl, devName)
	for _, rt := range vCfg.Routes {
		v.enableFwd(devName, "", vCfg.Subnet, rt)
		v.enableMasq(devName, "", vCfg.Subnet, rt)
	}
}

func (v *Switch) preNetVPN1(bridge, prefix string, vCfg *co.OpenVPN) {
	if vCfg == nil {
		return
	}
	// Enable MASQUERADE, and allowed forward.
	v.enableFwd("", bridge, vCfg.Subnet, prefix)
	v.enableMasq("", bridge, vCfg.Subnet, prefix)
}

func (v *Switch) preNets() {
	for _, nCfg := range v.cfg.Network {
		name := nCfg.Name
		w := NewNetworker(nCfg)
		v.worker[name] = w
		brCfg := nCfg.Bridge
		if brCfg == nil {
			continue
		}

		v.preWorker(w)
		brName := brCfg.Name
		vCfg := nCfg.OpenVPN

		ifAddr := strings.SplitN(brCfg.Address, "/", 2)[0]
		// Enable MASQUERADE for OpenVPN
		if vCfg != nil {
			v.preNetVPN0(nCfg, vCfg)
		}
		if ifAddr == "" {
			continue
		}
		subnet := w.Subnet()
		// Enable MASQUERADE, and allowed forward.
		for _, rt := range nCfg.Routes {
			v.preNetVPN1(brName, rt.Prefix, vCfg)
			v.enableFwd(brName, "", subnet, rt.Prefix)
			if rt.MultiPath != nil {
				v.enableSnat(brName, "", ifAddr, rt.Prefix)
			} else if rt.Mode == "snat" {
				v.enableMasq(brName, "", subnet, rt.Prefix)
			}
		}
	}
}

func (v *Switch) preApps() {
	// Append accessed auth for point
	v.apps.Auth = app.NewAccess(v)
	v.hooks = append(v.hooks, v.apps.Auth.OnFrame)
	// Append request process
	v.apps.Request = app.NewRequest(v)
	v.hooks = append(v.hooks, v.apps.Request.OnFrame)

	inspect := ""
	for _, v := range v.cfg.Inspect {
		inspect += v
	}
	// Check whether inspect neighbor
	if strings.Contains(inspect, "neighbor") {
		v.apps.Neighbor = app.NewNeighbors(v)
		v.hooks = append(v.hooks, v.apps.Neighbor.OnFrame)
	}
	// Check whether inspect online flow by five-tuple.
	if strings.Contains(inspect, "online") {
		v.apps.OnLines = app.NewOnline(v)
		v.hooks = append(v.hooks, v.apps.OnLines.OnFrame)
	}
	for i, h := range v.hooks {
		v.out.Debug("Switch.preApps: id %d, func %s", i, libol.FunName(h))
	}
}

func (v *Switch) preAcl() {
	for _, acl := range v.cfg.Acl {
		if acl.Name == "" {
			continue
		}
		v.firewall.AddChain(network.IpChain{
			Table: network.TRaw,
			Name:  acl.Name,
		})
		for _, rule := range acl.Rules {
			v.firewall.AddRule(network.IpRule{
				Table:   network.TRaw,
				Chain:   acl.Name,
				Source:  rule.SrcIp,
				Dest:    rule.DstIp,
				Proto:   rule.Proto,
				SrcPort: rule.SrcPort,
				DstPort: rule.DstPort,
				Jump:    rule.Action,
			})
		}
	}
}

func (v *Switch) GetPort(listen string) string {
	_, port := libol.GetHostPort(listen)
	return port
}

func (v *Switch) preAllowVPN(cfg *co.OpenVPN) {
	if cfg == nil {
		return
	}
	port := v.GetPort(cfg.Listen)
	if cfg.Protocol == "udp" {
		v.enablePort("udp", port)
	} else {
		v.enablePort("tcp", port)
	}
}

func (v *Switch) preAllow() {
	port := v.GetPort(v.cfg.Listen)
	UdpPorts := []string{"4500", "4500", "8472", "4789", port}
	TcpPorts := []string{"7471", port}
	if v.cfg.Http != nil {
		TcpPorts = append(TcpPorts, v.GetPort(v.cfg.Http.Listen))
	}
	v.enablePort("udp", strings.Join(UdpPorts, ","))
	v.enablePort("tcp", strings.Join(TcpPorts, ","))
	for _, nCfg := range v.cfg.Network {
		if nCfg.OpenVPN == nil {
			continue
		}
		v.preAllowVPN(nCfg.OpenVPN)
	}
}

func (v *Switch) SetPass(file string) {
	cache.User.SetFile(file)
}

func (v *Switch) LoadPass() {
	cache.User.Load()
}

func (v *Switch) Initialize() {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.preAcl()
	v.preAllow()
	v.preApps()
	if v.cfg.Http != nil {
		v.http = NewHttp(v)
	}
	v.preNets()
	// FireWall
	v.firewall.Initialize()
	for _, w := range v.worker {
		if w.Provider() == "vxlan" {
			continue
		}
		w.Initialize()
	}
	// Load password for guest access
	v.SetPass(v.cfg.PassFile)
	v.LoadPass()

	ldap := v.cfg.Ldap
	if ldap != nil {
		cache.User.SetLdap(&libol.LDAPConfig{
			Server:    ldap.Server,
			BindUser:  ldap.BindDN,
			BindPass:  ldap.BindPass,
			BaseDN:    ldap.BaseDN,
			Attr:      ldap.Attribute,
			Filter:    ldap.Filter,
			EnableTls: ldap.Tls,
		})
	}
	// Enable cert verify for access
	cert := v.cfg.Cert
	if cert != nil {
		cache.User.SetCert(&libol.CertConfig{
			Crt: cert.CrtFile,
		})
	}
	// Enable L2TP
	if v.cfg.L2TP != nil {
		v.l2tp = NewL2TP(v.cfg.L2TP)
	}
	// Start confd monitor
	v.confd.Initialize()
}

func (v *Switch) onFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	for _, h := range v.hooks {
		if v.out.Has(libol.LOG) {
			v.out.Log("Switch.onFrame: %s", libol.FunName(h))
		}
		if h != nil {
			if err := h(client, frame); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *Switch) OnClient(client libol.SocketClient) error {
	client.SetStatus(libol.ClConnected)
	v.out.Info("Switch.onClient: %s", client.String())
	return nil
}

func (v *Switch) SignIn(client libol.SocketClient) error {
	v.out.Cmd("Switch.SignIn %s", client.String())
	data := struct {
		Address string `json:"address"`
		Switch  string `json:"switch"`
	}{
		Address: client.String(),
		Switch:  client.LocalAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		v.out.Error("Switch.SignIn: %s", err)
		return err
	}
	v.out.Cmd("Switch.SignIn: %s", body)
	m := libol.NewControlFrame(libol.SignReq, body)
	if err := client.WriteMsg(m); err != nil {
		v.out.Error("Switch.SignIn: %s", err)
		return err
	}
	return nil
}

func client2Point(client libol.SocketClient) (*models.Point, error) {
	addr := client.RemoteAddr()
	if private := client.Private(); private == nil {
		return nil, libol.NewErr("point %s notFound.", addr)
	} else {
		obj, ok := private.(*models.Point)
		if !ok {
			return nil, libol.NewErr("point %s notRight.", addr)
		}
		return obj, nil
	}
}

func (v *Switch) ReadClient(client libol.SocketClient, frame *libol.FrameMessage) error {
	addr := client.RemoteAddr()
	if v.out.Has(libol.LOG) {
		v.out.Log("Switch.ReadClient: %s %x", addr, frame.Frame())
	}
	frame.Decode()
	if err := v.onFrame(client, frame); err != nil {
		v.out.Debug("Switch.ReadClient: %s dropping by %s", addr, err)
		if frame.Action() == libol.PingReq {
			// send sign message to point require login.
			_ = v.SignIn(client)
		}
		return nil
	}
	if frame.IsControl() {
		return nil
	}
	// process ethernet frame message.
	obj, err := client2Point(client)
	if err != nil {
		return err
	}
	device := obj.Device
	if device == nil {
		return libol.NewErr("Tap devices is nil")
	}
	if _, err := device.Write(frame.Frame()); err != nil {
		v.out.Error("Switch.ReadClient: %s", err)
		return err
	}
	return nil
}

func (v *Switch) OnClose(client libol.SocketClient) error {
	addr := client.RemoteAddr()
	v.out.Info("Switch.OnClose: %s", addr)
	if obj, err := client2Point(client); err == nil {
		cache.Network.DelLease(obj.Alias, obj.Network)
	}
	cache.Point.Del(addr)
	return nil
}

func (v *Switch) Start() {
	v.lock.Lock()
	defer v.lock.Unlock()

	OpenUDP()
	// firstly, start network.
	for _, w := range v.worker {
		if w.Provider() == "vxlan" {
			continue
		}
		w.Start(v)
	}
	// start server for accessing
	libol.Go(v.server.Accept)
	call := libol.ServerListener{
		OnClient: v.OnClient,
		OnClose:  v.OnClose,
		ReadAt:   v.ReadClient,
	}
	libol.Go(func() { v.server.Loop(call) })
	if v.http != nil {
		libol.Go(v.http.Start)
	}
	libol.Go(v.firewall.Start)
	libol.Go(v.confd.Start)
	if v.l2tp != nil {
		libol.Go(v.l2tp.Start)
	}
}

func (v *Switch) Stop() {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.out.Info("Switch.Stop")
	if v.l2tp != nil {
		v.l2tp.Stop()
	}
	v.confd.Stop()
	if v.http != nil {
		v.http.Shutdown()
	}
	// stop network.
	for _, w := range v.worker {
		if w.Provider() == "vxlan" {
			continue
		}
		w.Stop()
	}
	v.out.Info("Switch.Stop left points")
	// notify leave to point.
	for p := range cache.Point.List() {
		if p == nil {
			break
		}
		v.leftClient(p.Client)
	}
	v.firewall.Stop()
	v.server.Close()
}

func (v *Switch) Alias() string {
	return v.cfg.Alias
}

func (v *Switch) UpTime() int64 {
	return time.Now().Unix() - v.newTime
}

func (v *Switch) Server() libol.SocketServer {
	return v.server
}

func (v *Switch) GetBridge(tenant string) (network.Bridger, error) {
	w, ok := v.worker[tenant]
	if !ok {
		return nil, libol.NewErr("bridge %s notFound", tenant)
	}
	return w.Bridge(), nil
}

func (v *Switch) NewTap(tenant string) (network.Taper, error) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.out.Debug("Switch.NewTap")

	// already not need support free list for device.
	// dropped firstly packages during 15s because of forwarding delay.
	br, err := v.GetBridge(tenant)
	if err != nil {
		v.out.Error("Switch.NewTap: %s", err)
		return nil, err
	}
	dev, err := network.NewTaper(tenant, network.TapConfig{
		Provider: br.Type(),
		Type:     network.TAP,
		VirBuf:   v.cfg.Queue.VirWrt,
		KernBuf:  v.cfg.Queue.VirSnd,
		Name:     "auto",
	})
	if err != nil {
		v.out.Error("Switch.NewTap: %s", err)
		return nil, err
	}
	dev.Up()
	// add new tap to bridge.
	_ = br.AddSlave(dev.Name())
	v.out.Info("Switch.NewTap: %s on %s", dev.Name(), tenant)
	return dev, nil
}

func (v *Switch) FreeTap(dev network.Taper) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	name := dev.Name()
	tenant := dev.Tenant()
	v.out.Debug("Switch.FreeTap %s", name)
	w, ok := v.worker[tenant]
	if !ok {
		return libol.NewErr("bridge %s notFound", tenant)
	}
	br := w.Bridge()
	_ = br.DelSlave(dev.Name())
	v.out.Info("Switch.FreeTap: %s", name)
	return nil
}

func (v *Switch) UUID() string {
	if v.uuid == "" {
		v.uuid = libol.GenString(13)
	}
	return v.uuid
}

func (v *Switch) ReadTap(device network.Taper, readAt func(f *libol.FrameMessage) error) {
	name := device.Name()
	v.out.Info("Switch.ReadTap: %s", name)
	done := make(chan bool, 2)
	queue := make(chan *libol.FrameMessage, v.cfg.Queue.TapWr)
	libol.Go(func() {
		for {
			frame := libol.NewFrameMessage(0)
			n, err := device.Read(frame.Frame())
			if err != nil {
				v.out.Error("Switch.ReadTap: %s", err)
				done <- true
				break
			}
			frame.SetSize(n)
			if v.out.Has(libol.LOG) {
				v.out.Log("Switch.ReadTap: %x\n", frame.Frame()[:n])
			}
			queue <- frame
		}
	})
	defer device.Close()
	for {
		select {
		case frame := <-queue:
			if err := readAt(frame); err != nil {
				v.out.Error("Switch.ReadTap: readAt %s %s", name, err)
				return
			}
		case <-done:
			return
		}
	}
}

func (v *Switch) OffClient(client libol.SocketClient) {
	v.out.Info("Switch.OffClient: %s", client)
	if v.server != nil {
		v.server.OffClient(client)
	}
}

func (v *Switch) Config() *co.Switch {
	return co.Manager.Switch
}

func (v *Switch) leftClient(client libol.SocketClient) {
	if client == nil {
		return
	}
	v.out.Info("Switch.leftClient: %s", client.String())
	data := struct {
		DateTime   int64  `json:"datetime"`
		UUID       string `json:"uuid"`
		Alias      string `json:"alias"`
		Connection string `json:"connection"`
		Address    string `json:"address"`
	}{
		DateTime:   time.Now().Unix(),
		UUID:       v.UUID(),
		Alias:      v.Alias(),
		Address:    client.LocalAddr(),
		Connection: client.RemoteAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		v.out.Error("Switch.leftClient: %s", err)
		return
	}
	v.out.Cmd("Switch.leftClient: %s", body)
	m := libol.NewControlFrame(libol.LeftReq, body)
	if err := client.WriteMsg(m); err != nil {
		v.out.Error("Switch.leftClient: %s", err)
		return
	}
}

func (v *Switch) Firewall() *network.FireWall {
	return v.firewall
}

func (v *Switch) Reload() {
	co.Reload()
	cache.Reload()
	for _, w := range v.worker {
		w.Reload(v)
	}
	libol.Go(v.firewall.Start)
}

func (v *Switch) Save() {
	v.cfg.Save()
}
