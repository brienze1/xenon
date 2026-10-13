package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptReference struct {
	Members []ScriptMember `json:"members"`
}

func NewScriptReference(members []entity.ScriptMember) ScriptReference {
	result := ScriptReference{Members: make([]ScriptMember, 0, len(members))}
	for _, member := range members {
		result.Members = append(result.Members, NewScriptMember(member))
	}
	return result
}

func (s ScriptReference) LogValue() any {
	return map[string]any{"members": len(s.Members)}
}
