package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptSandboxLocation struct {
	Current string   `json:"current,omitempty"`
	Path    []string `json:"path"`
}

func NewScriptSandboxLocation(info entity.ScriptInfo) ScriptSandboxLocation {
	return ScriptSandboxLocation{Path: append([]string{}, info.Location...), Current: info.Current}
}
