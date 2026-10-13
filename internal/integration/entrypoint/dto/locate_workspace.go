package dto

import "path/filepath"

type LocateWorkspace struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Path    string `json:"path" validate:"required,absolute_path"`
}

func (l *LocateWorkspace) GetPath() string {
	return filepath.Clean(l.Path)
}
