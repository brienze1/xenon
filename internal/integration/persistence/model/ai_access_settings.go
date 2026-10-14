package model

import "github.com/brienze1/xenon/internal/domain/entity"

type AIAccessSettings struct {
	Enabled       *bool `json:"enabled,omitempty"`
	Port          int   `json:"port,omitempty"`
	SkipApprovals bool  `json:"skipApprovals"`
}

func NewAIAccessSettings(settings entity.AIAccessSettings) AIAccessSettings {
	return AIAccessSettings{Enabled: &settings.Enabled, Port: settings.Port, SkipApprovals: settings.SkipApprovals}
}

func (a *AIAccessSettings) ToEntity() entity.AIAccessSettings {
	access := entity.DefaultAIAccessSettings()
	if a.Enabled != nil {
		access.Enabled = *a.Enabled
	}
	access.Port = a.Port
	access.SkipApprovals = a.SkipApprovals
	return access
}
