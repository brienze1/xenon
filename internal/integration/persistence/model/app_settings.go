package model

import "github.com/brienze1/xenon/internal/domain/entity"

type AppSettings struct {
	AIAccess        *AIAccessSettings `json:"aiAccess,omitempty"`
	AutoUpdate      *bool             `json:"autoUpdate,omitempty"`
	ScriptTimeoutMs int               `json:"scriptTimeoutMs"`
	HistoryUserCap  int               `json:"historyUserCap,omitempty"`
	HistoryAICap    int               `json:"historyAiCap,omitempty"`
}

func NewAppSettings(settings entity.AppSettings) AppSettings {
	access := NewAIAccessSettings(settings.AIAccess)
	return AppSettings{
		ScriptTimeoutMs: settings.ScriptTimeoutMs,
		HistoryUserCap:  settings.HistoryUserCap,
		HistoryAICap:    settings.HistoryAICap,
		AIAccess:        &access,
		AutoUpdate:      &settings.AutoUpdate,
	}
}

func (a *AppSettings) ToEntity() entity.AppSettings {
	settings := entity.AppSettings{
		ScriptTimeoutMs: a.ScriptTimeoutMs,
		HistoryUserCap:  a.HistoryUserCap,
		HistoryAICap:    a.HistoryAICap,
		AIAccess:        entity.DefaultAIAccessSettings(),
		AutoUpdate:      a.AutoUpdate == nil || *a.AutoUpdate,
	}
	if a.AIAccess != nil {
		settings.AIAccess = a.AIAccess.ToEntity()
	}
	return settings.OrDefault()
}
