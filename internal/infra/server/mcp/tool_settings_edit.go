package mcp

import (
	"context"
	"encoding/json"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type settingsSaveArguments struct {
	appSettingsView
	AIAccess json.RawMessage `json:"ai_access"`
}

func (t *tools) settingsEdit(ctx context.Context, call Call) (Result, error) {
	var arguments settingsSaveArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	if arguments.AIAccess != nil {
		return Result{}, humanOnly("Changing AI access")
	}
	saved, err := t.controllers.Settings.Save(ctx, dto.AppSettings{ScriptTimeoutMs: arguments.ScriptTimeoutMs, HistoryUserCap: arguments.HistoryUserCap, HistoryAICap: arguments.HistoryAICap})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newAppSettingsView(*saved)}, nil
}
