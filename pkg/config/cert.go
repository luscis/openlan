package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/luscis/openlan/pkg/libol"
)

type Crypt struct {
	Algo   string `json:"algorithm,omitempty"`
	Secret string `json:"secret,omitempty"`
}

func (c *Crypt) IsZero() bool {
	return c.Algo == "" && c.Secret == ""
}

func (c *Crypt) Correct() {
	if c.Secret != "" && c.Algo == "" {
		c.Algo = "xor"
	}
}

type Cert struct {
	Dir      string `json:"directory"`
	CrtFile  string `json:"cert"`
	KeyFile  string `json:"key"`
	CaFile   string `json:"rootCa"`
	Insecure bool   `json:"insecure"`
}

func (c *Cert) Correct() {
	if c.Dir == "" {
		c.Dir = VarDir("cert")
	}
	if c.CrtFile == "" {
		c.CrtFile = fmt.Sprintf("%s/crt", c.Dir)
	}
	if c.KeyFile == "" {
		c.KeyFile = fmt.Sprintf("%s/key", c.Dir)
	}
	if c.CaFile == "" {
		c.CaFile = fmt.Sprintf("%s/ca-trusted.crt", c.Dir)
	}
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
