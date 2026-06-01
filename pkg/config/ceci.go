package config

import (
	"fmt"
	"strings"
)

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
	Name    string         `json:"-" yaml:"-"`
	Proxy   []*CeciProxy   `json:"proxy" yaml:"proxy"`
	Service []*CeciService `json:"service,omitempty" yaml:"service,omitempty"`
}

type CeciService struct {
	Name     string               `json:"-" yaml:"-"`
	Mode     string               `json:"mode,omitempty" yaml:"mode,omitempty"`
	Listen   string               `json:"listen" yaml:"listen"`
	Network  string               `json:"network,omitempty" yaml:"-"`
	Protocol string               `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Balance  string               `json:"balance,omitempty" yaml:"balance,omitempty"`
	Backends []string             `json:"backends,omitempty" yaml:"backends,omitempty"`
	Routes   []CeciServiceBackend `json:"routes,omitempty" yaml:"routes,omitempty"`
}

type CeciServiceBackend struct {
	Backends []string `json:"backends,omitempty" yaml:"backends,omitempty"`
	Match    []string `json:"match,omitempty" yaml:"match,omitempty"`
}

func (s *CeciService) Correct() {
	if s.Mode == "" {
		s.Mode = "service"
	}
	if s.Protocol == "" {
		s.Protocol = "tcp"
	}
	if s.Balance == "" {
		s.Balance = "roundrobin"
	}
}

func appendUnique(values []string, add ...string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values)+len(add))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	for _, value := range add {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func parseBackends(many []string) []string {
	backends := make([]string, 0, len(many))
	for _, item := range many {
		for _, part := range strings.FieldsFunc(item, func(r rune) bool { return r == '|' || r == ';' || r == ',' }) {
			part = strings.TrimSpace(part)
			if part != "" {
				backends = append(backends, part)
			}
		}
	}
	return appendUnique(nil, backends...)
}

func (s *CeciService) AddBackend(hostname string, backends []string) bool {
	values := parseBackends(backends)
	if len(values) == 0 {
		return false
	}
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		s.Backends = appendUnique(s.Backends, values...)
		return true
	}
	for i := range s.Routes {
		route := &s.Routes[i]
		for _, host := range route.Match {
			if strings.EqualFold(strings.TrimSpace(host), hostname) {
				route.Backends = appendUnique(route.Backends, values...)
				return true
			}
		}
	}
	s.Routes = append(s.Routes, CeciServiceBackend{
		Backends: values,
		Match:    []string{hostname},
	})
	return true
}

func (s *CeciSpecifies) Correct() {
	if s.Proxy == nil {
		s.Proxy = make([]*CeciProxy, 0)
	}
	if s.Service == nil {
		s.Service = make([]*CeciService, 0)
	}
	for _, t := range s.Proxy {
		t.Correct()
	}
	for _, t := range s.Service {
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

func (s *CeciService) Id() string {
	return fmt.Sprintf("%s", s.Listen)
}

func (s *CeciSpecifies) FindService(value *CeciService) (*CeciService, int) {
	for index, obj := range s.Service {
		if obj.Id() == value.Id() {
			return obj, index
		}
	}
	return nil, -1
}

func (s *CeciSpecifies) AddService(value *CeciService) bool {
	_, find := s.FindService(value)
	if find == -1 {
		s.Service = append(s.Service, value)
	}
	return find == -1
}

func (s *CeciSpecifies) DelService(value *CeciService) (*CeciService, bool) {
	obj, find := s.FindService(value)
	if find != -1 {
		s.Service = append(s.Service[:find], s.Service[find+1:]...)
	}
	return obj, find != -1
}

func (s *CeciSpecifies) ListProxy(call func(*CeciProxy)) {
	if s == nil || call == nil {
		return
	}
	for _, obj := range s.Proxy {
		call(obj)
	}
}

func (s *CeciSpecifies) ListService(call func(*CeciService)) {
	if s == nil || call == nil {
		return
	}
	for _, obj := range s.Service {
		call(obj)
	}
}
