package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptRunRequestOptions struct {
	Variables map[string]any `json:"variables"`
}

func (s *ScriptRunRequestOptions) ToEntity(scopes entity.ScriptScopes) entity.ScriptScopes {
	for name, value := range s.Variables {
		scopes.Local = scopes.Local.WithValue(name, value)
	}
	return scopes
}
