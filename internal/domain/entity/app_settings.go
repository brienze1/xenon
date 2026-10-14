package entity

import (
	"time"

	"github.com/brienze1/xenon/pkg/errs"
)

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

// Validate returns the first failing field in order: script timeout, user cap, AI cap, port.
func (a AppSettings) Validate() error {
	switch {
	case !a.HasValidScriptTimeout():
		return errs.BadRequestError("the script timeout must be between 1000 and 60000 ms", "SCR-01400",
			errs.ErrorDetails{Attribute: "script_timeout_ms", Messages: []string{"must be between 1000 and 60000"}})
	case !a.HasValidHistoryUserCap():
		return errs.BadRequestError("the History cap for your sends must be between 100 and 20000", "SET-02400",
			errs.ErrorDetails{Attribute: "history_user_cap", Messages: []string{"must be between 100 and 20000"}})
	case !a.HasValidHistoryAICap():
		return errs.BadRequestError("the History cap for AI activity must be between 100 and 20000", "SET-03400",
			errs.ErrorDetails{Attribute: "history_ai_cap", Messages: []string{"must be between 100 and 20000"}})
	case !a.HasValidAIPort():
		return errs.BadRequestError("the AI access port must be between 1024 and 65535", "SET-04400",
			errs.ErrorDetails{Attribute: "ai_access.port", Messages: []string{"must be between 1024 and 65535"}})
	default:
		return nil
	}
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

func (a AppSettings) SkipsAIApprovals() bool {
	return a.AIAccess.SkipApprovals
}

func (a AppSettings) Merged(patch AppSettingsPatch) AppSettings {
	return patch.ApplyTo(a)
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
