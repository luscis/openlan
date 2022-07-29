package api

import "os"

func GetEnv(key, value string) string {
	val := os.Getenv(key)
	if val == "" {
		return value
	}
	return val
}
