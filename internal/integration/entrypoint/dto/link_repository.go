package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/linkmode"
)

type LinkRepository struct {
	LocalId    string `json:"local_id" validate:"required,not_blank"`
	Repository string `json:"repository" validate:"required,not_blank"`
	Branch     string `json:"branch" validate:"required,not_blank"`
	Subfolder  string `json:"subfolder" validate:"safe_subfolder"`
	AccountID  string `json:"account_id" validate:"required,not_blank"`
	ClonePath  string `json:"clone_path" validate:"omitempty,absolute_path"`
	Mode       string `json:"mode" validate:"omitempty,oneof=adopt_repo replace_with_mine"`
}

func (l *LinkRepository) ToEntity() entity.SyncSettings {
	return entity.NewSyncSettings(l.Repository, l.Branch, l.Subfolder)
}

func (l *LinkRepository) LinkMode() enums.LinkMode {
	return linkmode.GetLinkMode(l.Mode)
}
