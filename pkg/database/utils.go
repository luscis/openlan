package database

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/ovsdb"
	"strconv"
	"strings"
)

func PrintError(result []ovsdb.OperationResult) {
	for _, ret := range result {
		if len(ret.Error) == 0 {
			continue
		}
		libol.Info("%s", ret.Details)
	}
}

func GenUUID() string {
	return libol.GenString(32)
}

func HasPrefix(value string, index int, dest string) bool {
	if len(value) >= index {
		return value[:index] == dest
	}
	return false
}

func GetAddrPort(conn string) (string, int) {
	values := strings.SplitN(conn, ":", 2)
	if len(values) == 2 {
		port, _ := strconv.Atoi(values[1])
		return values[0], port
	}
	return values[0], 0
}
