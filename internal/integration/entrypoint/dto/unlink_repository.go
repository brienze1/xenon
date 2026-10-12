package dto

import "path/filepath"

type UnlinkRepository struct {
	LocalId      string `json:"local_id" validate:"required,not_blank"`
	TargetFolder string `json:"target_folder" validate:"required,absolute_path"`
}

func (u *UnlinkRepository) GetTargetFolder() string {
	return filepath.Clean(u.TargetFolder)
}
