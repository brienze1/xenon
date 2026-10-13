package model

import "github.com/brienze1/xenon/internal/domain/entity"

type AppSettings struct {
	AIAccess        *AIAccessSettings `json:"aiAccess,omitempty"`
	ScriptTimeoutMs int               `json:"scriptTimeoutMs"`
	HistoryUserCap  int               `json:"historyUserCap,omitempty"`
	HistoryAICap    int               `json:"historyAiCap,omitempty"`
}

type AIAccessSettings struct {
	Enabled       *bool `json:"enabled,omitempty"`
	Port          int   `json:"port,omitempty"`
	SkipApprovals bool  `json:"skipApprovals"`
}

func NewAppSettings(settings entity.AppSettings) AppSettings {
	enabled := settings.AIAccess.Enabled
	return AppSettings{
		ScriptTimeoutMs: settings.ScriptTimeoutMs,
		HistoryUserCap:  settings.HistoryUserCap,
		HistoryAICap:    settings.HistoryAICap,
		AIAccess:        &AIAccessSettings{Enabled: &enabled, Port: settings.AIAccess.Port, SkipApprovals: settings.AIAccess.SkipApprovals},
	}
}

func (a *AppSettings) ToEntity() entity.AppSettings {
	settings := entity.AppSettings{
		ScriptTimeoutMs: a.ScriptTimeoutMs,
		HistoryUserCap:  a.HistoryUserCap,
		HistoryAICap:    a.HistoryAICap,
		AIAccess:        entity.DefaultAIAccessSettings(),
	}
	if a.AIAccess != nil {
		settings.AIAccess = a.AIAccess.ToEntity()
	}
	return settings.OrDefault()
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
