package model

import "github.com/brienze1/xenon/internal/domain/entity"

type AppSettings struct {
	ScriptTimeoutMs int `json:"scriptTimeoutMs"`
}

func NewAppSettings(settings entity.AppSettings) AppSettings {
	return AppSettings{ScriptTimeoutMs: settings.ScriptTimeoutMs}
}

func (a *AppSettings) ToEntity() entity.AppSettings {
	return entity.AppSettings{ScriptTimeoutMs: a.ScriptTimeoutMs}.OrDefault()
}
