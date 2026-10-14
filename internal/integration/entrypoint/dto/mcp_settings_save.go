package dto

import "encoding/json"

type MCPSettingsSave struct {
	ScriptTimeoutMs *int            `json:"script_timeout_ms,omitempty"`
	HistoryUserCap  *int            `json:"history_user_cap,omitempty"`
	HistoryAICap    *int            `json:"history_ai_cap,omitempty"`
	AIAccess        json.RawMessage `json:"ai_access"`
}

// ChangesAIAccess is true whenever the arguments name ai_access at all, even as null.
func (m *MCPSettingsSave) ChangesAIAccess() bool {
	return m.AIAccess != nil
}

func (m *MCPSettingsSave) AppSettings() AppSettings {
	return NewPartialAppSettings(m.ScriptTimeoutMs, m.HistoryUserCap, m.HistoryAICap)
}
