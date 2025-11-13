package config

import "fmt"

type CeciTcp struct {
	Name   string   `json:"-" yaml:"-"`
	Listen string   `json:"listen" yaml:"listen"`
	Target []string `json:"target,omitempty" yaml:"target,omitempty"`
}

func (s *CeciTcp) Correct() {
}

func (s *CeciTcp) Id() string {
	return fmt.Sprintf("%s", s.Listen)
}

type CeciHttp struct {
	Name   string `json:"-" yaml:"-"`
	Listen string `json:"listen" yaml:"listen"`
}

func (s *CeciHttp) Correct() {
}

func (s *CeciHttp) Id() string {
	return fmt.Sprintf("%s", s.Listen)
}

type CeciSpecifies struct {
	Name string      `json:"-" yaml:"-"`
	Tcp  []*CeciTcp  `json:"tcp" yaml:"tcp"`
	Http []*CeciHttp `json:"http" yaml:"http"`
}

func (s *CeciSpecifies) Correct() {
	if s.Tcp == nil {
		s.Tcp = make([]*CeciTcp, 0)
	}
	for _, t := range s.Tcp {
		t.Correct()
	}
	if s.Http == nil {
		s.Http = make([]*CeciHttp, 0)
	}
	for _, t := range s.Http {
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

func (s *CeciSpecifies) FindHttp(value *CeciHttp) (*CeciHttp, int) {
	for index, obj := range s.Http {
		if obj.Id() == value.Id() {
			return obj, index
		}
	}
	return nil, -1
}

func (s *CeciSpecifies) AddHttp(value *CeciHttp) bool {
	_, find := s.FindHttp(value)
	if find == -1 {
		s.Http = append(s.Http, value)
	}
	return find == -1
}

func (s *CeciSpecifies) DelHttp(value *CeciHttp) (*CeciHttp, bool) {
	obj, find := s.FindHttp(value)
	if find != -1 {
		s.Http = append(s.Http[:find], s.Http[find+1:]...)
	}
	return obj, find != -1
}
