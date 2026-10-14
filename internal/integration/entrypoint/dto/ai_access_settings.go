package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AIAccessSettings struct {
	Enabled       *bool `json:"enabled,omitempty"`
	Port          *int  `json:"port,omitempty"`
	SkipApprovals *bool `json:"skip_approvals,omitempty"`
}

func NewAIAccessSettings(settings entity.AIAccessSettings, port int) AIAccessSettings {
	return AIAccessSettings{Enabled: &settings.Enabled, Port: &port, SkipApprovals: &settings.SkipApprovals}
}

func (a *AIAccessSettings) ToEntity() entity.AIAccessPatch {
	return entity.AIAccessPatch{Enabled: a.Enabled, Port: a.Port, SkipApprovals: a.SkipApprovals}
}
