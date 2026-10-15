package entity

type AppSettingsPatch struct {
	ScriptTimeoutMs *int
	HistoryUserCap  *int
	HistoryAICap    *int
	AIAccess        *AIAccessPatch
	AutoUpdate      *bool
}

func (a AppSettingsPatch) ApplyTo(settings AppSettings) AppSettings {
	if a.ScriptTimeoutMs != nil {
		settings.ScriptTimeoutMs = *a.ScriptTimeoutMs
	}
	if a.HistoryUserCap != nil {
		settings.HistoryUserCap = *a.HistoryUserCap
	}
	if a.HistoryAICap != nil {
		settings.HistoryAICap = *a.HistoryAICap
	}
	if a.AutoUpdate != nil {
		settings.AutoUpdate = *a.AutoUpdate
	}
	settings.AIAccess = settings.AIAccess.Merged(a.AIAccess)
	return settings
}

func (a AppSettingsPatch) ChangesAIAccess() bool {
	return a.AIAccess != nil
}

func (a AppSettingsPatch) LogValue() any {
	return map[string]any{"script_timeout_ms": a.ScriptTimeoutMs, "history_user_cap": a.HistoryUserCap, "history_ai_cap": a.HistoryAICap, "ai_access": a.AIAccess, "auto_update": a.AutoUpdate}
}
