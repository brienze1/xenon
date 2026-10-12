package model

import "github.com/brienze1/xenon/internal/domain/entity"

type LocalSyncSettings struct {
	AccountID  string `json:"accountId"`
	ClonePath  string `json:"clonePath"`
	Repository string `json:"repository,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Subfolder  string `json:"subfolder,omitempty"`
	AutoPull   bool   `json:"autoPull"`
	AutoPush   bool   `json:"autoPush"`
}

func NewLocalSyncSettings(settings entity.LocalSyncSettings) LocalSyncSettings {
	return LocalSyncSettings{
		AccountID:  settings.AccountID,
		ClonePath:  settings.ClonePath,
		Repository: settings.Repository,
		Branch:     settings.Branch,
		Subfolder:  settings.Subfolder,
		AutoPull:   settings.AutoPull,
		AutoPush:   settings.AutoPush,
	}
}

func (l *LocalSyncSettings) ToEntity() entity.LocalSyncSettings {
	return entity.LocalSyncSettings{
		AccountID:  l.AccountID,
		ClonePath:  l.ClonePath,
		Repository: l.Repository,
		Branch:     l.Branch,
		Subfolder:  l.Subfolder,
		AutoPull:   l.AutoPull,
		AutoPush:   l.AutoPush,
	}
}
