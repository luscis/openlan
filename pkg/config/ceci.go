package config

import "fmt"

type CeciProxy struct {
	Name     string     `json:"-" yaml:"-"`
	Mode     string     `json:"mode" yaml:"mode"`
	Listen   string     `json:"listen" yaml:"listen"`
	Network  string     `json:"network,omitempty" yaml:"-"`
	Target   []string   `json:"target,omitempty" yaml:"target,omitempty"`
	Backends ToForwards `json:"backends,omitempty" yaml:"backends,omitempty"`
	Cert     *Cert      `json:"cert,omitempty" yaml:"cert,omitempty"`
}

func (s *CeciProxy) Correct() {
	if s.Mode == "" {
		s.Mode = "tcp"
	}
	if s.Cert != nil {
		s.Cert.Correct()
	}
}

func (s *CeciProxy) Id() string {
	return fmt.Sprintf("%s", s.Listen)
}

type CeciSpecifies struct {
	Name  string       `json:"-" yaml:"-"`
	Proxy []*CeciProxy `json:"proxy" yaml:"proxy"`
}

func (s *CeciSpecifies) Correct() {
	if s.Proxy == nil {
		s.Proxy = make([]*CeciProxy, 0)
	}
	for _, t := range s.Proxy {
		t.Correct()
	}
}

func (s *CeciSpecifies) FindProxy(value *CeciProxy) (*CeciProxy, int) {
	for index, obj := range s.Proxy {
		if obj.Id() == value.Id() {
			return obj, index
		}
	}
	return nil, -1
}

func (s *CeciSpecifies) AddProxy(value *CeciProxy) bool {
	_, find := s.FindProxy(value)
	if find == -1 {
		s.Proxy = append(s.Proxy, value)
	}
	return find == -1
}

func (s *CeciSpecifies) DelProxy(value *CeciProxy) (*CeciProxy, bool) {
	obj, find := s.FindProxy(value)
	if find != -1 {
		s.Proxy = append(s.Proxy[:find], s.Proxy[find+1:]...)
	}
	return obj, find != -1
}
