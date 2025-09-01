package config

import "fmt"

type BgpNeighbor struct {
	Name     string `json:"-" yaml:"-"`
	Address  string `json:"address" yaml:"address"`
	RemoteAs int    `json:"remoteas" yaml:"remoteas"`
	Secret   string `json:"secret,omitempty" yaml:"secret,omitempty"`
	State    string `json:"state,omitempty" yaml:"state,omitempty"`
}

func (s *BgpNeighbor) Correct() {
}

func (s *BgpNeighbor) Id() string {
	return fmt.Sprintf("%s", s.Address)
}

type BgpSpecifies struct {
	Name      string         `json:"-" yaml:"-"`
	LocalAs   int            `json:"localas" yaml:"localas"`
	RouteId   string         `json:"routeid" yaml:"routeid"`
	Neighbors []*BgpNeighbor `json:"neighbors" yaml:"neighbors"`
}

func (s *BgpSpecifies) Correct() {
	if s.Neighbors == nil {
		s.Neighbors = make([]*BgpNeighbor, 0)
	}
	for _, t := range s.Neighbors {
		t.Correct()
	}
}

func (s *BgpSpecifies) FindNeighbor(value *BgpNeighbor) (*BgpNeighbor, int) {
	for index, obj := range s.Neighbors {
		if obj.Id() == value.Id() {
			return obj, index
		}
	}
	return nil, -1
}

func (s *BgpSpecifies) AddNeighbor(value *BgpNeighbor) bool {
	_, find := s.FindNeighbor(value)
	if find == -1 {
		s.Neighbors = append(s.Neighbors, value)
	}
	return find == -1
}

func (s *BgpSpecifies) DelNeighbor(value *BgpNeighbor) (*BgpNeighbor, bool) {
	obj, find := s.FindNeighbor(value)
	if find != -1 {
		s.Neighbors = append(s.Neighbors[:find], s.Neighbors[find+1:]...)
	}
	return obj, find != -1
}
