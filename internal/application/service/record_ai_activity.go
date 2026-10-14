package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type recordAIActivity struct {
	findAppSettings      usecase.FindAppSettings
	findVault            usecase.FindVault
	registerKnownSecrets usecase.RegisterKnownSecrets
	saveHistoryEntry     usecase.SaveHistoryEntry
	publishAIEvent       usecase.PublishAIEvent
}

func NewRecordAIActivityService(
	findAppSettings usecase.FindAppSettings,
	findVault usecase.FindVault,
	registerKnownSecrets usecase.RegisterKnownSecrets,
	saveHistoryEntry usecase.SaveHistoryEntry,
	publishAIEvent usecase.PublishAIEvent,
) adapter.RecordAIActivityService {
	return &recordAIActivity{
		findAppSettings:      findAppSettings,
		findVault:            findVault,
		registerKnownSecrets: registerKnownSecrets,
		saveHistoryEntry:     saveHistoryEntry,
		publishAIEvent:       publishAIEvent,
	}
}

func (r *recordAIActivity) Record(ctx context.Context, localId, client string, activity entity.AIActivity) error {
	logger.Info(ctx, "Started", activity, map[string]any{"localId": localId, "client": client})

	if !entity.HasActiveWorkspace(localId) {
		logger.Info(ctx, "Finished", activity, map[string]any{"localId": localId, "client": client})
		return nil
	}
	if err := r.record(ctx, localId, client, activity); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", activity, map[string]any{"localId": localId, "client": client})
	return nil
}

func (r *recordAIActivity) record(ctx context.Context, localId, client string, activity entity.AIActivity) error {
	settings, err := r.findAppSettings.FindAppSettings(ctx)
	if err != nil {
		return err
	}
	vault, err := r.findVault.Find(ctx, localId)
	if err != nil {
		return err
	}
	known := r.registerKnownSecrets.Register(ctx, usecase.CollectKnownSecrets(vault)...)
	guarded := activity.WithArgs(usecase.MaskKnownSecrets(activity.RedactedArgs(), known))
	entry := entity.NewActionHistoryEntry(idutils.NewUlid(), timeutils.Now(), client, guarded)
	if err := r.saveHistoryEntry.Save(ctx, localId, entry, settings.HistoryCaps()); err != nil {
		return err
	}
	return r.publishAIEvent.HistoryChanged(ctx, localId)
}
