package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
)

// ScriptLogSink receives console calls while a script runs, so they reach the console in the order they happen.
type ScriptLogSink func(log ScriptLog)

// ScriptLog is one console call; Values keeps each argument as a JSON tree with special values tagged by "$type".
type ScriptLog struct {
	Scopes    *ScriptScopes
	Level     enums.ConsoleLevel
	Arguments []string
	Values    []any
}

func NewScriptLog(level enums.ConsoleLevel, arguments []string, values []any) ScriptLog {
	return ScriptLog{Level: level, Arguments: append([]string{}, arguments...), Values: append([]any{}, values...)}
}

// WithScopes records the variable scopes as the script had them when it logged.
func (s ScriptLog) WithScopes(scopes ScriptScopes) ScriptLog {
	s.Scopes = &scopes
	return s
}

func (s ScriptLog) Text() string {
	return strings.Join(s.Arguments, " ")
}

func (s ScriptLog) HasScopes() bool {
	return s.Scopes != nil
}

func (s ScriptLog) IsClear() bool {
	return s.Level == consolelevel.Clear
}
