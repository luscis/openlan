package config

import (
	"flag"
	"path/filepath"

	"github.com/luscis/openlan/pkg/libol"
)

type Perf struct {
	Access   int `json:"access" yaml:"access"`
	Neighbor int `json:"neighbor" yaml:"neighbor"`
	OnLine   int `json:"online" yaml:"online"`
	Link     int `json:"link" yaml:"link"`
	User     int `json:"user" yaml:"user"`
	Esp      int `json:"esp" yaml:"esp"`
	State    int `json:"state" yaml:"state"`
	Policy   int `json:"policy" yaml:"policy"`
	VxLAN    int `json:"vxlan" yaml:"vxlan"`
}

func (p *Perf) Correct() {
	if p.Access == 0 {
		p.Access = 64
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
	File      string              `json:"-" yaml:"-"`
	Alias     string              `json:"alias" yaml:"alias"`
	Perf      Perf                `json:"limit,omitempty" yaml:"limit,omitempty"`
	Protocol  string              `json:"protocol" yaml:"protocol"` // tcp, tls, udp, kcp, ws and wss.
	Listen    string              `json:"listen" yaml:"listen"`
	Timeout   int                 `json:"timeout" yaml:"timeout"`
	Http      *Http               `json:"http,omitempty" yaml:"http,omitempty"`
	Log       Log                 `json:"log" yaml:"log"`
	Cert      *Cert               `json:"cert,omitempty" yaml:"cert,omitempty"`
	Crypt     *Crypt              `json:"crypt,omitempty" yaml:"crypt,omitempty"`
	Network   map[string]*Network `json:"network,omitempty" yaml:"network,omitempty"`
	Acl       map[string]*ACL     `json:"acl,omitempty" yaml:"acl,omitempty"`
	Qos       map[string]*Qos     `json:"qos,omitempty" yaml:"qos,omitempty"`
	FireWall  []FlowRule          `json:"firewall,omitempty" yaml:"firewall,omitempty"`
	Queue     Queue               `json:"queue" yaml:"queue"`
	PassFile  string              `json:"password" yaml:"password"`
	Ldap      *LDAP               `json:"ldap,omitempty" yaml:"ldap,omitempty"`
	AddrPool  string              `json:"pool,omitempty" yaml:"pool,omitempty"`
	ConfDir   string              `json:"-" yaml:"-"`
	TokenFile string              `json:"-" yaml:"-"`
}

func NewSwitch() *Switch {
	s := &Switch{
		Acl:     make(map[string]*ACL, 32),
		Qos:     make(map[string]*Qos, 1024),
		Network: make(map[string]*Network, 32),
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

func (s *Switch) IsYaml() bool {
	return libol.IsYaml(s.File)
}

func (s *Switch) Initialize() {
	s.File = s.Dir("switch.json", "")
	if err := libol.FileExist(s.File); err != nil {
		s.File = s.Dir("switch.yaml", "")
	}
	if err := s.Load(); err != nil {
		libol.Error("Switch.Initialize %s", err)
	}
	s.Correct()
	s.LoadExtend()
	libol.Debug("Switch.Initialize %v", s)
}

func (s *Switch) LoadExtend() {
	s.LoadAcl()
	s.LoadQos()
	s.LoadNetworks()
}

func (s *Switch) Correct() {
	s.Log.Correct()
	s.Queue.Correct()

	if s.Alias == "" {
		s.Alias = GetAlias()
	}

	SetListen(&s.Listen, 10002)
	if s.Http == nil {
		s.Http = &Http{}
	}
	s.Http.Correct()

	vpn := DefaultOpenVPN()
	vpn.Url = s.Http.GetUrl()

	if s.Timeout == 0 {
		s.Timeout = 120
	}
	s.TokenFile = s.Dir("token", "")
	if s.Cert == nil {
		s.Cert = &Cert{}
	}
	s.Cert.Correct()

	if s.Crypt == nil {
		s.Crypt = &Crypt{}
	}
	s.Crypt.Correct()
	s.Perf.Correct()

	s.PassFile = s.Dir("password", "")
	if s.Protocol == "" {
		s.Protocol = "tcp"
	}
	if s.AddrPool == "" {
		s.AddrPool = "100.255"
	}
}

func (s *Switch) Dir(elem0, elem1 string) string {
	var file string

	if elem1 == "" {
		return filepath.Join(s.ConfDir, elem0)
	}

	if s.IsYaml() {
		file = elem1 + ".yaml"
	} else {
		file = elem1 + ".json"
	}

	return filepath.Join(s.ConfDir, elem0, file)
}

func (s *Switch) RemarshalNetwork(obj *Network, format string) {
	if obj.Specifies == nil {
		return
	}

	context := obj.Specifies
	obj.NewSpecifies()

	if format == "" {
		format = "json"
		if s.IsYaml() {
			format = "yaml"
		}
	}

	if format == "yaml" {
		if data, err := libol.MarshalYaml(context); err != nil {
			libol.Warn("Switch.Format %s", err)
		} else if err := libol.UnmarshalYaml(obj.Specifies, data); err != nil {
			libol.Warn("Switch.Format %s", err)
		}
	} else {
		if data, err := libol.Marshal(context, true); err != nil {
			libol.Warn("Switch.Format %s", err)
		} else if err := libol.Unmarshal(obj.Specifies, data); err != nil {
			libol.Warn("Switch.Format %s", err)
		}
	}
}

func (s *Switch) UnmarshalNetwork(data []byte) (*Network, error) {
	libol.Debug("Switch.UnmarshalNetwork %s", data)
	obj := &Network{
		Alias:   s.Alias,
		ConfDir: s.ConfDir,
	}
	if s.IsYaml() {
		if err := libol.UnmarshalYaml(obj, data); err != nil {
			return nil, err
		}
	} else {
		if err := libol.Unmarshal(obj, data); err != nil {
			return nil, err
		}
	}

	if _, ok := s.Network[obj.Name]; ok {
		return nil, libol.NewErr("already existed")
	}
	if obj.Bridge == nil {
		obj.Bridge = &Bridge{}
	}

	obj.LoadLink()
	obj.LoadRoute()
	obj.LoadOutput()
	obj.LoadFindHop()

	s.Network[obj.Name] = obj
	return obj, nil
}

func (s *Switch) RemarshalNetworks(format string) {
	for _, obj := range s.Network {
		s.RemarshalNetwork(obj, format)
	}
}

func (s *Switch) CorrectNetwork(obj *Network, format string) {
	s.RemarshalNetwork(obj, format)
	for _, link := range obj.Links {
		link.Correct()
	}
	obj.Correct(s)
	obj.Alias = s.Alias
	if obj.File == "" {
		obj.File = s.Dir("network", obj.Name)
	}
	if _, ok := s.Acl[obj.Name]; !ok {
		obj := &ACL{
			Name: obj.Name,
		}
		obj.Correct(s)
		s.Acl[obj.Name] = obj
	}
	if _, ok := s.Qos[obj.Name]; !ok {
		obj := &Qos{
			Name: obj.Name,
		}
		obj.Correct(s)
		s.Qos[obj.Name] = obj
	}
}

func (s *Switch) CorrectNetworks() {
	for _, obj := range s.Network {
		s.CorrectNetwork(obj, "")
	}
}

func (s *Switch) AddNetwork(data []byte) (*Network, error) {
	obj, err := s.UnmarshalNetwork(data)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *Switch) LoadNetworks() {
	files, err := filepath.Glob(s.Dir("network", "*"))
	if err != nil {
		libol.Error("Switch.LoadNetwork %s", err)
	}
	for _, file := range files {
		data, err := libol.LoadFile(file)
		if err != nil {
			libol.Warn("Switch.LoadNetwork %s", err)
			continue
		}
		if _, err := s.UnmarshalNetwork(data); err != nil {
			libol.Warn("Switch.LoadNetwork %s", err)
		}

	}
	s.CorrectNetworks()
}

func (s *Switch) LoadQos() {
	files, err := filepath.Glob(s.Dir("qos", "*"))
	if err != nil {
		libol.Error("Switch.LoadQos %s", err)
	}

	for _, file := range files {
		obj := &Qos{
			File: file,
		}
		if err := libol.UnmarshalLoad(obj, file); err != nil {
			libol.Error("Switch.LoadQos %s", err)
			continue
		}

		s.Qos[obj.Name] = obj
	}
	for _, obj := range s.Qos {
		for _, rule := range obj.Config {
			rule.Correct()
		}
		obj.Correct(s)
	}
}

func (s *Switch) LoadAcl() {
	files, err := filepath.Glob(s.Dir("acl", "*"))
	if err != nil {
		libol.Error("Switch.LoadAcl %s", err)
	}
	for _, file := range files {
		obj := &ACL{
			File: file,
		}
		if err := libol.UnmarshalLoad(obj, file); err != nil {
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
	return s.Network[name]
}

func (s *Switch) GetACL(name string) *ACL {
	return s.Acl[name]
}

func (s *Switch) GetQos(name string) *Qos {
	return s.Qos[name]
}
