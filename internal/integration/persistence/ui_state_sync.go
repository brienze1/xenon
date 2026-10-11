package persistence

import (
	"context"
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/logger"
)

const localSyncKey = "sync"

func (u *uiStatePersistence) FindSyncSettings(ctx context.Context, localId string) (*entity.LocalSyncSettings, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	u.mu.Lock()
	fields, err := readStateFields(localId)
	u.mu.Unlock()
	if err != nil {
		return nil, err
	}
	var stored model.LocalSyncSettings
	_ = json.Unmarshal(fields[localSyncKey], &stored)
	settings := stored.ToEntity()

	logger.Debug(ctx, "Finished", settings, map[string]any{"localId": localId})
	return &settings, nil
}

func (u *uiStatePersistence) SaveSyncSettings(ctx context.Context, localId string, settings entity.LocalSyncSettings) error {
	logger.Debug(ctx, "Started", settings, map[string]any{"localId": localId})

	u.mu.Lock()
	defer u.mu.Unlock()
	data, _ := json.Marshal(model.NewLocalSyncSettings(settings))
	if err := u.patch(localId, map[string]json.RawMessage{localSyncKey: data}, "UIS-12500", "UIS-13500"); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", settings, map[string]any{"localId": localId})
	return nil
}
