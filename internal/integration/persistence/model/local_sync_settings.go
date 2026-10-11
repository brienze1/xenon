package model

import "github.com/brienze1/xenon/internal/domain/entity"

type LocalSyncSettings struct {
	AccountID string `json:"accountId"`
	ClonePath string `json:"clonePath"`
	AutoPull  bool   `json:"autoPull"`
	AutoPush  bool   `json:"autoPush"`
}

func NewLocalSyncSettings(settings entity.LocalSyncSettings) LocalSyncSettings {
	return LocalSyncSettings{AccountID: settings.AccountID, ClonePath: settings.ClonePath, AutoPull: settings.AutoPull, AutoPush: settings.AutoPush}
}

func (l *LocalSyncSettings) ToEntity() entity.LocalSyncSettings {
	return entity.LocalSyncSettings{AccountID: l.AccountID, ClonePath: l.ClonePath, AutoPull: l.AutoPull, AutoPush: l.AutoPush}
}
