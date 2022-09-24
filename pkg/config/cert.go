package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/xtaci/kcp-go/v5"
	"io/ioutil"
)

type Crypt struct {
	Algo   string `json:"algo,omitempty" yaml:"algorithm"`
	Secret string `json:"secret,omitempty"`
}

func (c *Crypt) IsZero() bool {
	return c.Algo == "" && c.Secret == ""
}

func (c *Crypt) Default() {
	if c.Secret != "" && c.Algo == "" {
		c.Algo = "xor"
	}
}

type Cert struct {
	Dir      string `json:"dir" yaml:"directory"`
	CrtFile  string `json:"crt" yaml:"cert"`
	KeyFile  string `json:"key" yaml:"key"`
	CaFile   string `json:"ca" yaml:"rootCa"`
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
	caCert, err := ioutil.ReadFile(c.CaFile)
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

func GetBlock(cfg *Crypt) kcp.BlockCrypt {
	if cfg == nil || cfg.IsZero() {
		return nil
	}
	var block kcp.BlockCrypt
	pass := make([]byte, 64)
	if len(cfg.Secret) <= 64 {
		copy(pass, cfg.Secret)
	} else {
		copy(pass, []byte(cfg.Secret)[:64])
	}
	switch cfg.Algo {
	case "aes-128":
		block, _ = kcp.NewAESBlockCrypt(pass[:16])
	case "aes-192":
		block, _ = kcp.NewAESBlockCrypt(pass[:24])
	case "aes-256":
		block, _ = kcp.NewAESBlockCrypt(pass[:32])
	case "tea":
		block, _ = kcp.NewTEABlockCrypt(pass[:16])
	case "xtea":
		block, _ = kcp.NewXTEABlockCrypt(pass[:16])
	default:
		block, _ = kcp.NewSimpleXORBlockCrypt(pass)
	}
	return block
}
