package config

import (
	"fmt"
	"regexp"
	"strings"
)

func SplitSecret(value string) (string, string) {
	if strings.Contains(value, ":") {
		values := strings.SplitN(value, ":", 2)
		return values[0], values[1]
	}
	return value, ""
}

type ForwardSocks struct {
	Server string `json:"server,omitempty"`
}

type ForwardTo struct {
	Protocol string       `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Server   string       `json:"server,omitempty" yaml:"server,omitempty"`
	Insecure bool         `json:"insecure,omitempty" yaml:"insecure,omitempty"`
	Match    []string     `json:"match,omitempty" yaml:"match,omitempty"`
	Secret   string       `json:"secret,omitempty" yaml:"secret,omitempty"`
	Socks    ForwardSocks `json:"socks,omitempty" yaml:"socks,omitempty"`
	Nameto   string       `json:"nameto,omitempty" yaml:"nameto,omitempty"`
	CaCert   string       `json:"-" yaml:"-"`
}

func (f *ForwardTo) SocksAddr() string {
	if f.Socks.Server != "" {
		return f.Socks.Server
	}
	return f.Server
}

type ToForwards []*ForwardTo

func (h ToForwards) isMatch(value string, rules []string) bool {
	if len(rules) == 0 {
		return true
	}

	for _, rule := range rules {
		pattern := fmt.Sprintf(`(^|\.)%s(:\d+)?$`, regexp.QuoteMeta(rule))
		re := regexp.MustCompile(pattern)
		if re.MatchString(value) {
			return true
		}
	}

	return false
}

func (h ToForwards) FindBackend(host string) *ForwardTo {
	for _, via := range h {
		if via == nil {
			continue
		}
		if via.Server == "" && via.Socks.Server == "" {
			continue
		}
		if h.isMatch(host, via.Match) {
			return via
		}
	}

	return nil
}

type FindBackend interface {
	FindBackend(host string) *ForwardTo
}
