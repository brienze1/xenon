package dto

import "encoding/json"

type MCPSettingsSave struct {
	ScriptTimeoutMs *int            `json:"script_timeout_ms,omitempty" description:"Script time limit in milliseconds, 1000 to 60000."`
	HistoryUserCap  *int            `json:"history_user_cap,omitempty" description:"History entries kept for your sends, 100 to 20000."`
	HistoryAICap    *int            `json:"history_ai_cap,omitempty" description:"History entries kept for AI activity, 100 to 20000."`
	AIAccess        json.RawMessage `json:"ai_access" schema:"-"`
}

// ChangesAIAccess is true whenever the arguments name ai_access at all, even as null.
func (m *MCPSettingsSave) ChangesAIAccess() bool {
	return m.AIAccess != nil
}

func (m *MCPSettingsSave) AppSettings() AppSettings {
	return NewPartialAppSettings(m.ScriptTimeoutMs, m.HistoryUserCap, m.HistoryAICap)
}
