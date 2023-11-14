package schema

import "github.com/luscis/openlan/pkg/libol"

type Log struct {
	File  string `json:"file"`
	Level int    `json:"level"`
}

func NewLogSchema() Log {
	return Log{
		File:  libol.Logger.FileName,
		Level: libol.Logger.Level,
	}
}
