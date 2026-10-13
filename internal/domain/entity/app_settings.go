package entity

import "time"

const (
	DefaultScriptTimeoutMs = 10000
	MinScriptTimeoutMs     = 1000
	MaxScriptTimeoutMs     = 60000
)

type AppSettings struct {
	AIAccess        AIAccessSettings
	ScriptTimeoutMs int
	HistoryUserCap  int
	HistoryAICap    int
}

type AppSettingsPatch struct {
	ScriptTimeoutMs *int
	HistoryUserCap  *int
	HistoryAICap    *int
	AIAccess        *AIAccessPatch
}

func DefaultAppSettings() AppSettings {
	return AppSettings{
		ScriptTimeoutMs: DefaultScriptTimeoutMs,
		HistoryUserCap:  DefaultHistoryCap,
		HistoryAICap:    DefaultHistoryCap,
		AIAccess:        DefaultAIAccessSettings(),
	}
}

func (a AppSettings) HasValidScriptTimeout() bool {
	return a.ScriptTimeoutMs >= MinScriptTimeoutMs && a.ScriptTimeoutMs <= MaxScriptTimeoutMs
}

func (a AppSettings) HasValidHistoryUserCap() bool {
	return IsValidHistoryCap(a.HistoryUserCap)
}

func (a AppSettings) HasValidHistoryAICap() bool {
	return IsValidHistoryCap(a.HistoryAICap)
}

func (a AppSettings) HasValidAIPort() bool {
	return a.AIAccess.HasValidPort()
}

func (a AppSettings) OrDefault() AppSettings {
	defaults := DefaultAppSettings()
	if !a.HasValidScriptTimeout() {
		a.ScriptTimeoutMs = defaults.ScriptTimeoutMs
	}
	if !a.HasValidHistoryUserCap() {
		a.HistoryUserCap = defaults.HistoryUserCap
	}
	if !a.HasValidHistoryAICap() {
		a.HistoryAICap = defaults.HistoryAICap
	}
	if !a.HasValidAIPort() {
		a.AIAccess.Port = defaults.AIAccess.Port
	}
	return a
}

func (a AppSettings) Merged(patch AppSettingsPatch) AppSettings {
	if patch.ScriptTimeoutMs != nil {
		a.ScriptTimeoutMs = *patch.ScriptTimeoutMs
	}
	if patch.HistoryUserCap != nil {
		a.HistoryUserCap = *patch.HistoryUserCap
	}
	if patch.HistoryAICap != nil {
		a.HistoryAICap = *patch.HistoryAICap
	}
	a.AIAccess = a.AIAccess.Merged(patch.AIAccess)
	return a
}

func (a AppSettings) HistoryCaps() HistoryCaps {
	return HistoryCaps{User: a.HistoryUserCap, AI: a.HistoryAICap}.OrDefault()
}

func (a AppSettings) ScriptTimeout() time.Duration {
	return time.Duration(a.ScriptTimeoutMs) * time.Millisecond
}

func (a AppSettings) LogValue() any {
	return map[string]any{
		"script_timeout_ms": a.ScriptTimeoutMs,
		"history_user_cap":  a.HistoryUserCap,
		"history_ai_cap":    a.HistoryAICap,
		"ai_access":         a.AIAccess.LogValue(),
	}
}

func (a AppSettingsPatch) LogValue() any {
	return map[string]any{"script_timeout_ms": a.ScriptTimeoutMs, "history_user_cap": a.HistoryUserCap, "history_ai_cap": a.HistoryAICap, "ai_access": a.AIAccess}
}

func (a AppSettingsPatch) ChangesAIAccess() bool {
	return a.AIAccess != nil
}
