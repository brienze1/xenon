package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncLinkSettings struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Subfolder  string `json:"subfolder"`
	AccountID  string `json:"account_id"`
	ClonePath  string `json:"clone_path"`
	AutoPull   bool   `json:"auto_pull"`
	AutoPush   bool   `json:"auto_push"`
}

func NewSyncLinkSettings(settings entity.LocalSyncSettings) SyncLinkSettings {
	return SyncLinkSettings{
		Repository: settings.Repository,
		Branch:     settings.Branch,
		Subfolder:  settings.Subfolder,
		AccountID:  settings.AccountID,
		ClonePath:  settings.ClonePath,
		AutoPull:   settings.AutoPull,
		AutoPush:   settings.AutoPush,
	}
}
