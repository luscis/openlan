package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
)

type Crypt struct {
	Algo   string `json:"algorithm,omitempty" yaml:"algorithm,omitempty"`
	Secret string `json:"secret,omitempty" yaml:"secret,omitempty"`
}

func (c *Crypt) IsZero() bool {
	return c.Algo == "" && c.Secret == ""
}

func (c *Crypt) Correct() {
	if c.Secret != "" && c.Algo == "" {
		c.Algo = "xor"
	}
}

func (c *Crypt) Short() string {
	return c.Algo + ":" + c.Secret
}

type Cert struct {
	Dir      string `json:"-" yaml:"-"`
	CrtFile  string `json:"cert" yaml:"cert"`
	KeyFile  string `json:"key" yaml:"key"`
	CaFile   string `json:"rootCa" yaml:"rootCa"`
	CrtData  string `json:"certData,omitempty" yaml:"certData,omitempty"`
	KeyData  string `json:"keyData,omitempty" yaml:"keyData,omitempty"`
	CaData   string `json:"rootCaData,omitempty" yaml:"rootCaData,omitempty"`
	Insecure bool   `json:"insecure" yaml:"insecure"`
}

func (c *Cert) Correct() {
	if c.Dir == "" {
		c.Dir = VarDir("cert")
	}
	if c.CrtFile == "" && c.CrtData == "" {
		c.CrtFile = fmt.Sprintf("%s/crt", c.Dir)
	}
	if c.KeyFile == "" && c.KeyData == "" {
		c.KeyFile = fmt.Sprintf("%s/key", c.Dir)
	}
	if c.CaFile == "" && c.CaData == "" {
		c.CaFile = fmt.Sprintf("%s/ca.crt", c.Dir)
	}
}

func (c *Cert) LoadData() error {
	if c == nil {
		return nil
	}
	if c.CrtData == "" && strings.TrimSpace(c.CrtFile) != "" {
		data, err := os.ReadFile(c.CrtFile)
		if err != nil {
			return err
		}
		c.CrtData = string(data)
	}
	if c.KeyData == "" && strings.TrimSpace(c.KeyFile) != "" {
		data, err := os.ReadFile(c.KeyFile)
		if err != nil {
			return err
		}
		c.KeyData = string(data)
	}
	if c.CaData == "" && strings.TrimSpace(c.CaFile) != "" {
		data, err := os.ReadFile(c.CaFile)
		if err != nil {
			return err
		}
		c.CaData = string(data)
	}
	return nil
}

func (c *Cert) GetCertificates() []tls.Certificate {
	if c.KeyFile == "" || c.CrtFile == "" {
		return nil
	}
	libol.Debug("Cert.GetCertificates: %v", c)
	cer, err := tls.LoadX509KeyPair(c.CrtFile, c.KeyFile)
	if err != nil {
		libol.Error("Cert.GetCertificates: %s", err)
		return nil
	}
	return []tls.Certificate{cer}
}

func GetCertPool(ca string) (*x509.CertPool, error) {
	if ca == "" {
		return nil, libol.NewErr("%s: not such file", ca)
	}
	if err := libol.FileExist(ca); err != nil {
		return nil, libol.NewErr("Cert.GetTlsCertPool: %s not such file", ca)
	}
	caCert, err := os.ReadFile(ca)
	if err != nil {
		return nil, libol.NewErr("Cert.GetTlsCertPool: %s", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, libol.NewErr("Cert.GetTlsCertPool: invalid cert")
	}
	return pool, nil
}

func (c *Cert) GetCertPool() *x509.CertPool {
	pool, err := GetCertPool(c.CaFile)
	if err != nil {
		libol.Warn("GetCertPool %s", err)
		return nil
	}
	return pool
}
