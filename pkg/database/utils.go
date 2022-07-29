package database

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/ovn-org/libovsdb/ovsdb"
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
	return libol.GenRandom(32)
}
