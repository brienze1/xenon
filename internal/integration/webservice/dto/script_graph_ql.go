package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptGraphQl struct {
	Variables     any    `json:"variables,omitempty"`
	Query         string `json:"query"`
	OperationName string `json:"operationName,omitempty"`
}

func NewScriptGraphQl(body entity.BodyGraphQl) ScriptGraphQl {
	graphQl := ScriptGraphQl{Query: body.Query}
	if body.Variables != "" {
		graphQl.Variables = body.Variables
	}
	return graphQl
}

func (s *ScriptGraphQl) Equals(other *ScriptGraphQl) bool {
	if s == nil || other == nil {
		return s == other
	}
	return s.Query == other.Query && s.OperationName == other.OperationName && scriptText(s.Variables) == scriptText(other.Variables)
}

func (s *ScriptGraphQl) ToEntity() entity.BodyGraphQl {
	return entity.BodyGraphQl{Query: s.Query, Variables: scriptText(s.Variables)}
}
