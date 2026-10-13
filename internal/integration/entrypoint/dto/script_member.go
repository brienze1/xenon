package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptMember struct {
	Path        string `json:"path"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

func NewScriptMember(member entity.ScriptMember) ScriptMember {
	return ScriptMember{Path: member.Path, Signature: member.Signature, Description: member.Description, Kind: member.Kind.String()}
}
