package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptmemberkind"
)

type ScriptMember struct {
	Path        string `json:"path"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

func (s ScriptMember) ToEntity() entity.ScriptMember {
	return entity.ScriptMember{Path: s.Path, Signature: s.Signature, Description: s.Description, Kind: scriptmemberkind.GetScriptMemberKind(s.Kind)}
}
