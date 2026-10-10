package entity

import "time"

const (
	DefaultScriptTimeoutMs = 10000
	MinScriptTimeoutMs     = 1000
	MaxScriptTimeoutMs     = 60000
)

type AppSettings struct {
	ScriptTimeoutMs int
}

func DefaultAppSettings() AppSettings {
	return AppSettings{ScriptTimeoutMs: DefaultScriptTimeoutMs}
}

func (a AppSettings) HasValidScriptTimeout() bool {
	return a.ScriptTimeoutMs >= MinScriptTimeoutMs && a.ScriptTimeoutMs <= MaxScriptTimeoutMs
}

func (a AppSettings) OrDefault() AppSettings {
	if a.HasValidScriptTimeout() {
		return a
	}
	return DefaultAppSettings()
}

func (a AppSettings) ScriptTimeout() time.Duration {
	return time.Duration(a.ScriptTimeoutMs) * time.Millisecond
}

func (a AppSettings) LogValue() any {
	return map[string]any{"script_timeout_ms": a.ScriptTimeoutMs}
}
