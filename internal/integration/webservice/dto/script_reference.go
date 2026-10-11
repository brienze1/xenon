package dto

import "github.com/brienze1/xenon/internal/domain/entity"

// ScriptReference is the generated reference.json: the sandbox surface with signatures, descriptions and kinds.
type ScriptReference struct {
	Members []ScriptMember `json:"members"`
}

func (s ScriptReference) ToEntity() []entity.ScriptMember {
	members := make([]entity.ScriptMember, 0, len(s.Members))
	for _, member := range s.Members {
		members = append(members, member.ToEntity())
	}
	return members
}
