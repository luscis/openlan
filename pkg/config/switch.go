package config

import (
	"flag"
	"path/filepath"

	"github.com/luscis/openlan/pkg/libol"
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
	File      string          `json:"file"`
	Alias     string          `json:"alias"`
	Perf      Perf            `json:"limit,omitempty"`
	Protocol  string          `json:"protocol"` // tcp, tls, udp, kcp, ws and wss.
	Listen    string          `json:"listen"`
	Timeout   int             `json:"timeout"`
	Http      *Http           `json:"http,omitempty"`
	Log       Log             `json:"log"`
	Cert      *Cert           `json:"cert,omitempty"`
	Crypt     *Crypt          `json:"crypt,omitempty"`
	Network   []*Network      `json:"network,omitempty"`
	Acl       map[string]*ACL `json:"acl,omitempty"`
	Qos       map[string]*Qos `json:"qos,omitempty"`
	FireWall  []FlowRule      `json:"firewall,omitempty"`
	Queue     Queue           `json:"queue"`
	PassFile  string          `json:"password"`
	Ldap      *LDAP           `json:"ldap,omitempty"`
	AddrPool  string          `json:"pool,omitempty"`
	ConfDir   string          `json:"-"`
	TokenFile string          `json:"-"`
	L2TP      *L2TP           `json:"l2tp"`
}

func NewSwitch() *Switch {
	s := &Switch{
		Acl: make(map[string]*ACL, 32),
		Qos: make(map[string]*Qos, 1024),
	}
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
	s.LoadQos()
	s.LoadNetwork()
}

func (s *Switch) Correct() {
	s.Log.Correct()
	s.Queue.Correct()

	if s.Alias == "" {
		s.Alias = GetAlias()
	}

	CorrectAddr(&s.Listen, 10002)
	if s.Http == nil {
		s.Http = &Http{}
	}
	s.Http.Correct()

	vpn := DefaultOpenVPN()
	vpn.Url = s.Http.GetUrl()

	if s.Timeout == 0 {
		s.Timeout = 120
	}

	s.TokenFile = s.Dir("token")
	s.File = s.Dir("switch.json")

	if s.Cert == nil {
		s.Cert = &Cert{}
	}
	s.Cert.Correct()

	if s.Crypt == nil {
		s.Crypt = &Crypt{}
	}
	s.Crypt.Correct()
	s.Perf.Correct()

	s.PassFile = s.Dir("password")
	if s.Protocol == "" {
		s.Protocol = "tcp"
	}
	if s.AddrPool == "" {
		s.AddrPool = "169.254"
	}
}

func (s *Switch) Dir(elem ...string) string {
	args := append([]string{s.ConfDir}, elem...)
	return filepath.Join(args...)
}

func (s *Switch) Format() {
	for _, obj := range s.Network {
		context := obj.Specifies
		obj.NewSpecifies()
		if obj.Specifies == nil {
			continue
		}
		if data, err := libol.Marshal(context, true); err != nil {
			libol.Warn("Switch.Format %s", err)
		} else if err := libol.Unmarshal(obj.Specifies, data); err != nil {
			libol.Warn("Switch.Format %s", err)
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
		obj.Correct(s)
		obj.Alias = s.Alias
		if obj.File == "" {
			obj.File = s.Dir("network", obj.Name+".json")
		}
		if _, ok := s.Acl[obj.Name]; !ok {
			obj := &ACL{
				Name: obj.Name,
			}
			obj.Correct(s)
			s.Acl[obj.Name] = obj
		}
	}
}

func (s *Switch) LoadQos() {
	files, err := filepath.Glob(s.Dir("qos", "*.json"))
	if err != nil {
		libol.Error("Switch.LoadQos %s", err)
	}

	for _, k := range files {
		obj := &Qos{
			File: k,
		}
		if err := libol.UnmarshalLoad(obj, k); err != nil {
			libol.Error("Switch.LoadQos %s", err)
			continue
		}

		s.Qos[obj.Name] = obj
	}
	for _, obj := range s.Qos {
		for _, rule := range obj.Config {
			rule.Correct()
		}
		if obj.File == "" {
			obj.File = s.Dir("acl", obj.Name+".json")
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
		obj.Correct(s)
		s.Acl[obj.Name] = obj
	}
}

func (s *Switch) Load() error {
	return libol.UnmarshalLoad(s, s.File)
}

func (s *Switch) Save() {
	tmp := *s
	tmp.Acl = nil
	tmp.Qos = nil
	tmp.Network = nil
	if err := libol.MarshalSave(&tmp, tmp.File, true); err != nil {
		libol.Error("Switch.Save %s", err)
	}
	s.SaveAcl()
	s.SaveQos()
	s.SaveNetwork()
}

func (s *Switch) SaveQos() {
	for _, obj := range s.Qos {
		obj.Save()
	}
}

func (s *Switch) SaveAcl() {
	for _, obj := range s.Acl {
		obj.Save()
	}
}

func (s *Switch) SaveNetwork() {
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

func (s *Switch) GetACL(name string) *ACL {
	return s.Acl[name]
}

func (s *Switch) GetQos(name string) *Qos {
	return s.Qos[name]
}
