package cswitch

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/luscis/openlan/pkg/api"
	"github.com/luscis/openlan/pkg/app"
	"github.com/luscis/openlan/pkg/cache"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/network"
	cn "github.com/luscis/openlan/pkg/network"
	"github.com/luscis/openlan/pkg/schema"
)

const (
	UDPBin = "openudp"
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
			c.Tls = &tls.Config{
				Certificates: s.Cert.GetCertificates(),
			}
		}
		return libol.NewTcpServer(s.Listen, c)
	}
}

type Apps struct {
	Auth    *app.Access
	Request *app.Request
}

type Hook func(client libol.SocketClient, frame *libol.FrameMessage) error

type Switch struct {
	lock    sync.Mutex
	cfg     *co.Switch
	apps    Apps
	fire    *network.FireWallGlobal
	hooks   []Hook
	http    *Http
	server  libol.SocketServer
	worker  map[string]api.NetworkApi
	uuid    string
	newTime int64
	out     *libol.SubLogger
}

func NewSwitch(c *co.Switch) *Switch {
	server := GetSocketServer(c)
	v := &Switch{
		cfg:     c,
		fire:    network.NewFireWallGlobal(c.FireWall),
		worker:  make(map[string]api.NetworkApi, 32),
		server:  server,
		newTime: time.Now().Unix(),
		hooks:   make([]Hook, 0, 64),
		out:     libol.NewSubLogger(c.Alias),
	}
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
	v.fire.AddRule(network.IPRule{
		Table:   network.TFilter,
		Chain:   network.OLCInput,
		Proto:   protocol,
		Match:   "multiport",
		DstPort: port,
		Comment: "Open Default Ports",
	})
}

func (v *Switch) AddNetwork(network string) {
	for _, nCfg := range v.cfg.Network {
		name := nCfg.Name
		if name == network {
			w := NewNetworker(nCfg)
			v.worker[name] = w
			w.Initialize()
			w.Start(v)
		}
	}
}

func (v *Switch) DelNetwork(network string) {
	worker := v.worker[network]
	file := worker.Config().File

	worker.Stop(true)
	cache.Network.Del(network)
	delete(v.worker, network)
	delete(v.cfg.Network, network)
	if err := os.Remove(file); err != nil {
		v.out.Error("Error removing file: %s, err: %s", file, err)
	}
}

func (v *Switch) SaveNetwork(network string) {
	if network == "" {
		for _, obj := range v.cfg.Network {
			obj.Save()
		}
	} else {
		if obj := v.cfg.GetNetwork(network); obj != nil {
			obj.Save()
		}
	}
}

func (v *Switch) preNetwork() {
	for _, nCfg := range v.cfg.Network {
		name := nCfg.Name
		w := NewNetworker(nCfg)
		v.worker[name] = w
	}
}

func (v *Switch) preApplication() {
	// Append accessed auth for Access
	v.apps.Auth = app.NewAccess(v)
	v.hooks = append(v.hooks, v.apps.Auth.OnFrame)
	// Append request process
	v.apps.Request = app.NewRequest(v)
	v.hooks = append(v.hooks, v.apps.Request.OnFrame)

	for i, h := range v.hooks {
		v.out.Debug("Switch.preApplication: id %d, func %s", i, libol.FunName(h))
	}
}

func (v *Switch) GetPort(listen string) string {
	_, port := libol.GetHostPort(listen)
	return port
}

func (v *Switch) openPorts() {
	v.fire.AddRule(cn.IPRule{
		Table:   network.TFilter,
		Chain:   network.OLCForward,
		CtState: "RELATED,ESTABLISHED",
		Comment: "Accept Related",
	})
	port := v.GetPort(v.cfg.Listen)
	UdpPorts := []string{"500", "4500", "8472", "4789", port}
	TcpPorts := []string{port}
	if v.cfg.Http != nil {
		TcpPorts = append(TcpPorts, v.GetPort(v.cfg.Http.Listen))
	}
	v.enablePort("udp", strings.Join(UdpPorts, ","))
	v.enablePort("tcp", strings.Join(TcpPorts, ","))
}

func (v *Switch) Initialize() {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.openPorts()
	v.preApplication()
	if v.cfg.Http != nil {
		v.http = NewHttp(v)
	}
	v.preNetwork()
	// Load global firewall
	v.fire.Initialize()
	for _, w := range v.worker {
		w.Initialize()
	}
	// Load password for guest access
	cache.User.SetFile(v.cfg.PassFile)
	cache.User.Load()
	ldap := v.cfg.Ldap
	if ldap != nil {
		cfg := &libol.LDAPConfig{
			Server:    ldap.Server,
			BindUser:  ldap.BindDN,
			BindPass:  ldap.BindPass,
			BaseDN:    ldap.BaseDN,
			Attr:      ldap.Attribute,
			Filter:    ldap.Filter,
			EnableTls: ldap.Tls,
		}
		cache.User.SetLDAP(cfg)
	}

	// Enable cert verify for access
	cert := v.cfg.Cert
	if cert != nil {
		cache.User.SetCert(&libol.CertConfig{
			Crt: cert.CrtFile,
		})
	}
}

func (v *Switch) UpdateCert(data schema.VersionCert) {
	cert := v.cfg.Cert
	if cert == nil {
		return
	}

	value := data.Cert
	if value != "" {
		if err := os.WriteFile(cert.CrtFile, []byte(value), 0600); err != nil {
			v.out.Warn("Switch.UpdateCert: %s", err)
		}
	}

	value = data.Key
	if value != "" {
		if err := os.WriteFile(cert.KeyFile, []byte(value), 0600); err != nil {
			v.out.Warn("Switch.UpdateCert: %s", err)
		} else {
			v.out.Info("Switch.UpdateCert: please restart for cert key")
		}
	}

	value = data.Ca
	if value != "" {
		if err := os.WriteFile(cert.CaFile, []byte(value), 0600); err != nil {
			v.out.Warn("Switch.UpdateCert: %s", err)
		} else {
			v.out.Info("Switch.UpdateCert: please restart for cert ca")
		}
	}
}

func (v *Switch) GetCert() (ce schema.VersionCert) {
	cert := v.cfg.Cert
	if cert == nil {
		return ce
	}

	certData, err := os.ReadFile(cert.CrtFile)
	if err != nil {
		return ce
	}
	ce.Cert = string(certData)
	block, rest := pem.Decode(certData)
	if block == nil || len(rest) > 0 {
		return ce
	}
	xcert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return ce
	}
	ce.CertExpire = xcert.NotAfter.Format(time.RFC3339)

	certData, err = os.ReadFile(cert.CaFile)
	if err != nil {
		return ce
	}
	ce.Ca = string(certData)
	block, rest = pem.Decode(certData)
	if block == nil || len(rest) > 0 {
		return ce
	}
	xcert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return ce
	}
	ce.CaExpire = xcert.NotAfter.Format(time.RFC3339)

	return ce
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

func client2Access(client libol.SocketClient) (*models.Access, error) {
	addr := client.RemoteAddr()
	if private := client.Private(); private == nil {
		return nil, libol.NewErr("Access %s notFound.", addr)
	} else {
		obj, ok := private.(*models.Access)
		if !ok {
			return nil, libol.NewErr("Access %s notRight.", addr)
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
			// send sign message to Access require login.
			_ = v.SignIn(client)
		}
		return nil
	}
	if frame.IsControl() {
		return nil
	}
	// process ethernet frame message.
	obj, err := client2Access(client)
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
	if obj, err := client2Access(client); err == nil {
		cache.Network.DelLease(obj.Alias, obj.Network)
	}
	cache.Access.Del(addr)
	return nil
}

func (v *Switch) Start() {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.fire.Start()
	// firstly, start network.
	for _, w := range v.worker {
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
}

func (v *Switch) Stop() {
	v.lock.Lock()
	defer v.lock.Unlock()

	v.out.Info("Switch.Stop")

	if v.http != nil {
		v.http.Shutdown()
	}
	// stop network.
	for _, w := range v.worker {
		w.Stop(false)
	}
	v.out.Info("Switch.Stop left access")
	// notify leave to access.
	for p := range cache.Access.List() {
		if p == nil {
			break
		}
		v.leftClient(p.Client)
	}
	v.server.Close()
	//v.fire.Stop()
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
	return w.Bridger(), nil
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
	br := w.Bridger()
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
	return v.cfg
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

func (v *Switch) Save() {
	v.cfg.Save()
}

func (v *Switch) AddRate(device string, mbit int) {
	rate := fmt.Sprintf("%dMbit", mbit)
	burst := "64Kb"
	latency := "400ms"

	// Egress limit.
	out, err := libol.Exec("tc", "qdisc", "add", "dev", device, "root",
		"tbf", "rate", rate, "burst", burst, "latency", latency)
	if err != nil {
		v.out.Warn("Switch.AddRate: %s %d %s", device, mbit, out)
	}

	// Ingress limit.
	out, err = libol.Exec("tc", "qdisc", "add", "dev", device, "ingress")
	if err != nil {
		v.out.Warn("Switch.AddRate: %s %s", device, out)
	}
	out, err = libol.Exec("tc", "filter", "add", "dev", device, "parent", "ffff:",
		"protocol", "ip", "prio", "1", "u32", "match", "u32", "0", "0",
		"police", "rate", rate, "burst", burst, "drop", "flowid", ":1")
	if err != nil {
		v.out.Warn("Switch.AddRate: %s %d %s", device, mbit, out)
	}
}

func (v *Switch) DelRate(device string) {
	out, err := libol.Exec("tc", "qdisc", "del", "dev", device, "root")
	if err != nil {
		v.out.Debug("Switch.AddRate: %s %s", device, out)
	}
	out, err = libol.Exec("tc", "qdisc", "del", "dev", device, "ingress")
	if err != nil {
		v.out.Debug("Switch.AddRate: %s %s", device, out)
	}
}

func (v *Switch) AddLDAP(value schema.LDAP) error {
	v.cfg.Ldap = &co.LDAP{
		Server:    value.Server,
		BindDN:    value.BindDN,
		BindPass:  value.BindPass,
		BaseDN:    value.BaseDN,
		Filter:    value.Filter,
		Attribute: value.Attribute,
		Tls:       value.EnableTls,
	}
	ldap := v.cfg.Ldap
	cfg := &libol.LDAPConfig{
		Server:    ldap.Server,
		BindUser:  ldap.BindDN,
		BindPass:  ldap.BindPass,
		BaseDN:    ldap.BaseDN,
		Attr:      ldap.Attribute,
		Filter:    ldap.Filter,
		EnableTls: ldap.Tls,
	}
	cache.User.SetLDAP(cfg)
	return nil
}

func (v *Switch) DelLDAP() {
	v.cfg.Ldap = nil
	cache.User.ClearLDAP()
}
