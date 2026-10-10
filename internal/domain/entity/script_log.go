package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type ScriptLog struct {
	Level     enums.ConsoleLevel
	Arguments []string
}

func (s ScriptLog) Text() string {
	return strings.Join(s.Arguments, " ")
}
