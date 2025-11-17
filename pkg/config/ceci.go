package config

import "fmt"

type CeciTcp struct {
	Name   string   `json:"-" yaml:"-"`
	Mode   string   `json:"mode" yaml:"mode"`
	Listen string   `json:"listen" yaml:"listen"`
	Target []string `json:"target,omitempty" yaml:"target,omitempty"`
}

func (s *CeciTcp) Correct() {
	if s.Mode == "" {
		s.Mode = "tcp"
	}
}

func (s *CeciTcp) Id() string {
	return fmt.Sprintf("%s", s.Listen)
}

type CeciSpecifies struct {
	Name string     `json:"-" yaml:"-"`
	Tcp  []*CeciTcp `json:"tcp" yaml:"tcp"`
}

func (s *CeciSpecifies) Correct() {
	if s.Tcp == nil {
		s.Tcp = make([]*CeciTcp, 0)
	}
	for _, t := range s.Tcp {
		t.Correct()
	}
}

func (s *CeciSpecifies) FindTcp(value *CeciTcp) (*CeciTcp, int) {
	for index, obj := range s.Tcp {
		if obj.Id() == value.Id() {
			return obj, index
		}
	}
	return nil, -1
}

func (s *CeciSpecifies) AddTcp(value *CeciTcp) bool {
	_, find := s.FindTcp(value)
	if find == -1 {
		s.Tcp = append(s.Tcp, value)
	}
	return find == -1
}

func (s *CeciSpecifies) DelTcp(value *CeciTcp) (*CeciTcp, bool) {
	obj, find := s.FindTcp(value)
	if find != -1 {
		s.Tcp = append(s.Tcp[:find], s.Tcp[find+1:]...)
	}
	return obj, find != -1
}
