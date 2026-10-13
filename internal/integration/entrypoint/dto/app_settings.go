package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AppSettings struct {
	ScriptTimeoutMs int `json:"script_timeout_ms" validate:"min=1000,max=60000"`
}

func NewAppSettings(settings entity.AppSettings) AppSettings {
	return AppSettings{ScriptTimeoutMs: settings.ScriptTimeoutMs}
}

func (a *AppSettings) ToEntity() entity.AppSettings {
	return entity.AppSettings{ScriptTimeoutMs: a.ScriptTimeoutMs}
}
