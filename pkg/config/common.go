package config

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/luscis/openlan/pkg/libol"
)

var index = 99

func GenName(prefix string) string {
	index += 1
	return fmt.Sprintf("%s%d", prefix, index)
}

func VarDir(name ...string) string {
	return "/var/openlan/" + strings.Join(name, "/")
}

type Log struct {
	File    string `json:"file,omitempty"`
	Verbose int    `json:"level,omitempty"`
}

func (l *Log) Correct() {
	if l.Verbose == 0 {
		l.Verbose = libol.INFO
	}
}

func LogFile(file string) string {
	if runtime.GOOS == "linux" {
		return "/var/log/" + file
	}
	return file
}

type Http struct {
	Listen string `json:"listen,omitempty"`
	Public string `json:"public,omitempty"`
}

func (h *Http) Correct() {
	SetListen(&h.Listen, 10000)
	if h.Public == "" {
		h.Public = VarDir("public")
	}
}

func (h *Http) GetUrl() string {
	port := "10000"
	values := strings.SplitN(h.Listen, ":", 2)
	if len(values) == 2 {
		port = values[1]
	}
	return "https://127.0.0.1:" + port
}

func SetListen(listen *string, port int) {
	if *listen == "" {
		*listen = fmt.Sprintf("0.0.0.0:%d", port)
		return
	}
	values := strings.SplitN(*listen, ":", 2)
	if len(values) == 1 {
		*listen = fmt.Sprintf("%s:%d", values[0], port)
	}
}

func GetAlias() string {
	if hostname, err := os.Hostname(); err == nil {
		return strings.ToLower(hostname)
	}
	return libol.GenString(13)
}
