package config

import (
	"flag"
	"github.com/luscis/openlan/pkg/libol"
	"path/filepath"
)

type Perf struct {
	Point    int `json:"point"`
	Neighbor int `json:"neighbor"`
	OnLine   int `json:"online"`
	Link     int `json:"link"`
	User     int `json:"user"`
	Esp      int `json:"esp"`
	State    int `json:"state"`
	Policy   int `json:"policy"`
	VxLAN    int `json:"vxlan"`
}

func (p *Perf) Correct() {
	if p.Point == 0 {
		p.Point = 64
	}
	if p.Neighbor == 0 {
		p.Neighbor = 64
	}
	if p.OnLine == 0 {
		p.OnLine = 64
	}
	if p.Link == 0 {
		p.Link = 64
	}
	if p.User == 0 {
		p.User = 1024
	}
	if p.Esp == 0 {
		p.Esp = 64
	}
	if p.State == 0 {
		p.State = 64 * 4
	}
	if p.Policy == 0 {
		p.Policy = 64 * 8
	}
	if p.VxLAN == 0 {
		p.VxLAN = 64
	}
}

type Switch struct {
	File      string     `json:"file"`
	Alias     string     `json:"alias"`
	Perf      Perf       `json:"limit,omitempty"`
	Protocol  string     `json:"protocol"` // tcp, tls, udp, kcp, ws and wss.
	Listen    string     `json:"listen"`
	Timeout   int        `json:"timeout"`
	Http      *Http      `json:"http,omitempty"`
	Log       Log        `json:"log"`
	Cert      *Cert      `json:"cert,omitempty"`
	Crypt     *Crypt     `json:"crypt,omitempty"`
	Network   []*Network `json:"network,omitempty"`
	Acl       []*ACL     `json:"acl,omitempty"`
	FireWall  []FlowRule `json:"firewall,omitempty"`
	Inspect   []string   `json:"inspect,omitempty"`
	Queue     Queue      `json:"queue"`
	PassFile  string     `json:"password"`
	Ldap      *LDAP      `json:"ldap,omitempty"`
	AddrPool  string     `json:"pool,omitempty"`
	ConfDir   string     `json:"-"`
	TokenFile string     `json:"-"`
}

func NewSwitch() *Switch {
	s := Manager.Switch
	s.Parse()
	s.Initialize()
	return s
}

func (s *Switch) Parse() {
	flag.StringVar(&s.Log.File, "log:file", "", "Configure log file")
	flag.StringVar(&s.ConfDir, "conf:dir", "", "Configure switch's directory")
	flag.IntVar(&s.Log.Verbose, "log:level", 20, "Configure log level")
	flag.Parse()
}

func (s *Switch) Initialize() {
	s.File = s.Dir("switch.json")
	if err := s.Load(); err != nil {
		libol.Error("Switch.Initialize %s", err)
	}
	s.Correct()
	s.LoadExt()
	libol.Debug("Switch.Initialize %v", s)
}

func (s *Switch) LoadExt() {
	s.LoadAcl()
	s.LoadNetwork()
}

func (s *Switch) Correct() {
	if s.Alias == "" {
		s.Alias = GetAlias()
	}
	if s.Listen == "" {
		s.Listen = "0.0.0.0:10002"
	}
	CorrectAddr(&s.Listen, 10002)
	if s.Http == nil {
		s.Http = &Http{
			Listen: "0.0.0.0:10000",
		}
	}
	if s.Http != nil {
		CorrectAddr(&s.Http.Listen, 10000)
	}
	if s.Timeout == 0 {
		s.Timeout = 120
	}
	libol.Debug("Proxy.Correct Http %v", s.Http)
	s.TokenFile = filepath.Join(s.ConfDir, "token")
	s.File = filepath.Join(s.ConfDir, "switch.json")
	if s.Cert == nil {
		s.Cert = &Cert{}
	}
	s.Cert.Correct()
	if s.Crypt == nil {
		s.Crypt = &Crypt{}
	}
	s.Log.Correct()
	s.Crypt.Correct()
	s.Perf.Correct()
	s.PassFile = filepath.Join(s.ConfDir, "password")
	if s.Protocol == "" {
		s.Protocol = "tcp"
	}
	if s.AddrPool == "" {
		s.AddrPool = "100.44"
	}
	s.Queue.Correct()
}

func (s *Switch) Dir(elem ...string) string {
	args := append([]string{s.ConfDir}, elem...)
	return filepath.Join(args...)
}

func (s *Switch) Format() {
	for _, obj := range s.Network {
		libol.Debug("Switch.Format %s", obj)
		context := obj.Specifies
		switch obj.Provider {
		case "esp":
			obj.Specifies = &ESPSpecifies{}
		case "vxlan":
			obj.Specifies = &VxLANSpecifies{}
		case "fabric":
			obj.Specifies = &FabricSpecifies{}
		default:
			obj.Specifies = nil
			continue
		}
		if data, err := libol.Marshal(context, true); err == nil {
			if err := libol.Unmarshal(obj.Specifies, data); err != nil {
				libol.Warn("Switch.Format %s", err)
			}
		}
	}
}

func (s *Switch) LoadNetwork() {
	files, err := filepath.Glob(s.Dir("network", "*.json"))
	if err != nil {
		libol.Error("Switch.LoadNetwork %s", err)
	}
	for _, k := range files {
		obj := &Network{
			Alias:   s.Alias,
			File:    k,
			ConfDir: s.ConfDir,
		}
		if err := libol.UnmarshalLoad(obj, k); err != nil {
			libol.Error("Switch.LoadNetwork %s", err)
			continue
		}
		obj.LoadLink()
		obj.LoadRoute()
		s.Network = append(s.Network, obj)
	}
	s.Format()
	for _, obj := range s.Network {
		for _, link := range obj.Links {
			link.Correct()
		}
		obj.Correct()
		obj.Alias = s.Alias
		if obj.File == "" {
			obj.File = s.Dir("network", obj.Name+".json")
		}
	}
}

func (s *Switch) LoadAcl() {
	files, err := filepath.Glob(s.Dir("acl", "*.json"))
	if err != nil {
		libol.Error("Switch.LoadAcl %s", err)
	}
	for _, k := range files {
		obj := &ACL{
			File: k,
		}
		if err := libol.UnmarshalLoad(obj, k); err != nil {
			libol.Error("Switch.LoadAcl %s", err)
			continue
		}
		s.Acl = append(s.Acl, obj)
	}
	for _, obj := range s.Acl {
		for _, rule := range obj.Rules {
			rule.Correct()
		}
		if obj.File == "" {
			obj.File = s.Dir("acl", obj.Name+".json")
		}
	}
}

func (s *Switch) Load() error {
	return libol.UnmarshalLoad(s, s.File)
}

func (s *Switch) Save() {
	tmp := *s
	tmp.Acl = nil
	tmp.Network = nil
	if err := libol.MarshalSave(&tmp, tmp.File, true); err != nil {
		libol.Error("Switch.Save %s", err)
	}
	s.SaveAcl()
	s.SaveNetwork()
}

func (s *Switch) SaveAcl() {
	if s.Acl == nil {
		return
	}
	for _, obj := range s.Acl {
		if err := libol.MarshalSave(obj, obj.File, true); err != nil {
			libol.Error("Switch.Save.Acl %s %s", obj.Name, err)
		}
	}
}

func (s *Switch) SaveNetwork() {
	if s.Network == nil {
		return
	}
	for _, obj := range s.Network {
		obj.Save()
	}
}

func (s *Switch) Reload() {
	for _, obj := range s.Network {
		obj.Reload()
	}
}

func (s *Switch) GetNetwork(name string) *Network {
	for _, obj := range s.Network {
		if obj.Name == name {
			return obj
		}
	}
	return nil
}
