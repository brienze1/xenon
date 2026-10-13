package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type appSettingsView struct {
	ScriptTimeoutMs *int `json:"script_timeout_ms,omitempty"`
	HistoryUserCap  *int `json:"history_user_cap,omitempty"`
	HistoryAICap    *int `json:"history_ai_cap,omitempty"`
}

func newAppSettingsView(settings dto.AppSettings) appSettingsView {
	return appSettingsView{ScriptTimeoutMs: settings.ScriptTimeoutMs, HistoryUserCap: settings.HistoryUserCap, HistoryAICap: settings.HistoryAICap}
}

func (t *tools) settings(ctx context.Context, _ Call) (Result, error) {
	settings, err := t.controllers.Settings.Get(ctx)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newAppSettingsView(*settings)}, nil
}
