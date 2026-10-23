package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type FileFilter struct {
	DisplayName string `json:"display_name" validate:"required,not_blank"`
	Pattern     string `json:"pattern" validate:"required,not_blank"`
}

func (f *FileFilter) ToEntity() entity.FileFilter {
	return entity.FileFilter{DisplayName: f.DisplayName, Pattern: f.Pattern}
}
