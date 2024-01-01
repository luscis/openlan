package api

import (
	"os"
	"strings"
)

func GetEnv(key, value string) string {
	val := os.Getenv(key)
	if val == "" {
		return value
	}
	return val
}

func SplitName(name string) (string, string) {
	values := strings.SplitN(name, "@", 2)
	if len(values) == 2 {
		return values[0], values[1]
	}
	return name, ""
}

func SplitSocket(value string) (string, string) {
	values := strings.SplitN(value, ":", 2)
	if len(values) == 2 {
		return values[0], values[1]
	}
	return value, ""
}
