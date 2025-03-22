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

func (c *Cert) GetTlsCfg() *tls.Config {
	if c.KeyFile == "" || c.CrtFile == "" {
		return nil
	}
	libol.Debug("Cert.GetTlsCfg: %v", c)
	cer, err := tls.LoadX509KeyPair(c.CrtFile, c.KeyFile)
	if err != nil {
		libol.Error("Cert.GetTlsCfg: %s", err)
		return nil
	}
	return &tls.Config{Certificates: []tls.Certificate{cer}}
}

func (c *Cert) GetCertPool() *x509.CertPool {
	if c.CaFile == "" {
		return nil
	}
	if err := libol.FileExist(c.CaFile); err != nil {
		libol.Debug("Cert.GetTlsCertPool: %s not such file", c.CaFile)
		return nil
	}
	caCert, err := os.ReadFile(c.CaFile)
	if err != nil {
		libol.Warn("Cert.GetTlsCertPool: %s", err)
		return nil
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		libol.Warn("Cert.GetTlsCertPool: invalid cert")
	}
	return pool
}
