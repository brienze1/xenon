package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPAppSettings struct {
	ScriptTimeoutMs *int `json:"script_timeout_ms,omitempty"`
	HistoryUserCap  *int `json:"history_user_cap,omitempty"`
	HistoryAICap    *int `json:"history_ai_cap,omitempty"`
}

func NewMCPAppSettings(settings entity.AppSettings) MCPAppSettings {
	return MCPAppSettings{ScriptTimeoutMs: &settings.ScriptTimeoutMs, HistoryUserCap: &settings.HistoryUserCap, HistoryAICap: &settings.HistoryAICap}
}

func (m MCPAppSettings) LogValue() any {
	return map[string]any{"script_timeout_ms": m.ScriptTimeoutMs, "history_user_cap": m.HistoryUserCap, "history_ai_cap": m.HistoryAICap}
}
