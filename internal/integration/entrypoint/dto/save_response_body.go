package dto

import "path/filepath"

type SaveResponseBody struct {
	TabId string `json:"tab_id" validate:"required,not_blank"`
	Path  string `json:"path" validate:"required,absolute_path"`
}

func (s *SaveResponseBody) GetPath() string {
	return filepath.Clean(s.Path)
}
