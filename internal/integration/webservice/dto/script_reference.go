package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptmemberkind"
	"github.com/brienze1/xenon/internal/domain/enums/snippetphase"
)

// ScriptReference is the generated reference.json: the sandbox surface with signatures, descriptions and kinds.
type ScriptReference struct {
	Members []ScriptMember `json:"members"`
}

type ScriptMember struct {
	Path        string `json:"path"`
	Signature   string `json:"signature"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

type ScriptSnippet struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Phase string `json:"phase"`
	Text  string `json:"text"`
}

func (s ScriptReference) ToEntity() []entity.ScriptMember {
	members := make([]entity.ScriptMember, 0, len(s.Members))
	for _, member := range s.Members {
		members = append(members, entity.ScriptMember{Path: member.Path, Signature: member.Signature, Description: member.Description, Kind: scriptmemberkind.GetScriptMemberKind(member.Kind)})
	}
	return members
}

func (s ScriptSnippet) ToEntity() entity.ScriptSnippet {
	return entity.ScriptSnippet{ID: s.Id, Title: s.Title, Phase: snippetphase.GetSnippetPhase(s.Phase), Text: s.Text}
}
