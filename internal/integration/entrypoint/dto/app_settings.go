package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AppSettings struct {
	ScriptTimeoutMs *int              `json:"script_timeout_ms,omitempty"`
	HistoryUserCap  *int              `json:"history_user_cap,omitempty"`
	HistoryAICap    *int              `json:"history_ai_cap,omitempty"`
	AIAccess        *AIAccessSettings `json:"ai_access,omitempty"`
}

type AIAccessSettings struct {
	Enabled       *bool `json:"enabled,omitempty"`
	Port          *int  `json:"port,omitempty"`
	SkipApprovals *bool `json:"skip_approvals,omitempty"`
}

func NewAppSettings(settings entity.AppSettings, aiPortOverride *int) AppSettings {
	port := settings.AIAccess.EffectivePort(aiPortOverride)
	return AppSettings{
		ScriptTimeoutMs: &settings.ScriptTimeoutMs,
		HistoryUserCap:  &settings.HistoryUserCap,
		HistoryAICap:    &settings.HistoryAICap,
		AIAccess:        &AIAccessSettings{Enabled: &settings.AIAccess.Enabled, Port: &port, SkipApprovals: &settings.AIAccess.SkipApprovals},
	}
}

func (a *AppSettings) ToEntity() entity.AppSettingsPatch {
	patch := entity.AppSettingsPatch{ScriptTimeoutMs: a.ScriptTimeoutMs, HistoryUserCap: a.HistoryUserCap, HistoryAICap: a.HistoryAICap}
	if a.AIAccess != nil {
		patch.AIAccess = &entity.AIAccessPatch{Enabled: a.AIAccess.Enabled, Port: a.AIAccess.Port, SkipApprovals: a.AIAccess.SkipApprovals}
	}
	return patch
}

func (a AppSettings) LogValue() any {
	return map[string]any{"script_timeout_ms": a.ScriptTimeoutMs, "history_user_cap": a.HistoryUserCap, "history_ai_cap": a.HistoryAICap, "ai_access": a.AIAccess}
}
