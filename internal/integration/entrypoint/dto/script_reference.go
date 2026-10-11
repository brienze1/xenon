package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptReference struct {
	Members []ScriptMember `json:"members"`
}

type ScriptMember struct {
	Path        string `json:"path"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

func NewScriptReference(members []entity.ScriptMember) ScriptReference {
	result := ScriptReference{Members: make([]ScriptMember, 0, len(members))}
	for _, member := range members {
		result.Members = append(result.Members, ScriptMember{Path: member.Path, Signature: member.Signature, Description: member.Description, Kind: member.Kind.String()})
	}
	return result
}

func (s ScriptReference) LogValue() any {
	return map[string]any{"members": len(s.Members)}
}
