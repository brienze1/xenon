package dto

import "github.com/brienze1/xenon/internal/domain/entity"

func NewSaveSyncSettings(localId string, settings entity.LocalSyncSettings) SaveSyncSettings {
	return SaveSyncSettings{LocalId: localId, AccountID: settings.AccountID, AutoPull: settings.AutoPull, AutoPush: settings.AutoPush}
}
