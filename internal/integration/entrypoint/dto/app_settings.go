package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AppSettings struct {
	ScriptTimeoutMs *int              `json:"script_timeout_ms,omitempty"`
	HistoryUserCap  *int              `json:"history_user_cap,omitempty"`
	HistoryAICap    *int              `json:"history_ai_cap,omitempty"`
	AIAccess        *AIAccessSettings `json:"ai_access,omitempty"`
}

func NewAppSettings(settings entity.AppSettings, aiPortOverride *int) AppSettings {
	access := NewAIAccessSettings(settings.AIAccess, settings.AIAccess.EffectivePort(aiPortOverride))
	return AppSettings{
		ScriptTimeoutMs: &settings.ScriptTimeoutMs,
		HistoryUserCap:  &settings.HistoryUserCap,
		HistoryAICap:    &settings.HistoryAICap,
		AIAccess:        &access,
	}
}

func NewPartialAppSettings(scriptTimeoutMs, historyUserCap, historyAICap *int) AppSettings {
	return AppSettings{ScriptTimeoutMs: scriptTimeoutMs, HistoryUserCap: historyUserCap, HistoryAICap: historyAICap}
}

func (a *AppSettings) ToEntity() entity.AppSettingsPatch {
	patch := entity.AppSettingsPatch{ScriptTimeoutMs: a.ScriptTimeoutMs, HistoryUserCap: a.HistoryUserCap, HistoryAICap: a.HistoryAICap}
	if a.AIAccess != nil {
		access := a.AIAccess.ToEntity()
		patch.AIAccess = &access
	}
	return patch
}

func (a AppSettings) LogValue() any {
	return map[string]any{"script_timeout_ms": a.ScriptTimeoutMs, "history_user_cap": a.HistoryUserCap, "history_ai_cap": a.HistoryAICap, "ai_access": a.AIAccess}
}
