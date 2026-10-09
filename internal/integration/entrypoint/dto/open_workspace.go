package dto

import "path/filepath"

type OpenWorkspace struct {
	Path string `json:"path" validate:"required,absolute_path"`
}

func (o *OpenWorkspace) GetPath() string {
	return filepath.Clean(o.Path)
}
