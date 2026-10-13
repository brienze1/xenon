package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type CloneWorkspace struct {
	Repository string `json:"repository" validate:"required,not_blank"`
	Branch     string `json:"branch"`
	Subfolder  string `json:"subfolder" validate:"safe_subfolder"`
	AccountID  string `json:"account_id" validate:"required,not_blank"`
	ClonePath  string `json:"clone_path" validate:"omitempty,absolute_path"`
	Name       string `json:"name" validate:"max=200"`
}

func (c *CloneWorkspace) ToEntity() entity.SyncSettings {
	return entity.NewSyncSettings(c.Repository, c.Branch, c.Subfolder)
}
