package config

import (
	"fmt"
	"github.com/luscis/openlan/pkg/libol"
	"os"
	"runtime"
	"strings"
)

var index = 1024

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

func LogFile(file string) string {
	if runtime.GOOS == "linux" {
		return "/var/log/" + file
	}
	return file
}

type Http struct {
	Listen string `json:"listen,omitempty"`
	Public string `json:"public,omitempty" yaml:"publicDir"`
}

func CorrectAddr(listen *string, port int) {
	values := strings.Split(*listen, ":")
	if len(values) == 1 {
		*listen = fmt.Sprintf("%s:%d", values[0], port)
	}
}

func GetAlias() string {
	if hostname, err := os.Hostname(); err == nil {
		return strings.ToLower(hostname)
	}
	return libol.GenRandom(13)
}
