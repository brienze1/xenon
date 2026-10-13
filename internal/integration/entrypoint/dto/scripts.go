package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

type Scripts struct {
	PreRequest   string   `json:"pre_request"`
	PostResponse string   `json:"post_response"`
	Disabled     []string `json:"disabled" validate:"dive,oneof=preRequest postResponse"`
}

func NewScripts(scripts entity.Scripts) *Scripts {
	result := &Scripts{PreRequest: scripts.PreRequest, PostResponse: scripts.PostResponse, Disabled: make([]string, 0, len(scripts.Disabled))}
	for _, phase := range scripts.Disabled {
		result.Disabled = append(result.Disabled, phase.String())
	}
	return result
}

func (s *Scripts) ToEntity() entity.Scripts {
	scripts := entity.Scripts{PreRequest: s.PreRequest, PostResponse: s.PostResponse, Disabled: nil}
	for _, raw := range s.Disabled {
		scripts.Disabled = append(scripts.Disabled, scriptphase.GetScriptPhase(raw))
	}
	return scripts
}

func inheritsScripts(inherit *bool) bool {
	return inherit == nil || *inherit
}
