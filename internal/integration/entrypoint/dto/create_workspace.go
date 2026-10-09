package dto

import "path/filepath"

type CreateWorkspace struct {
	Name string `json:"name" validate:"required,not_blank,max=200"`
	Path string `json:"path" validate:"required,absolute_path"`
}

func (c *CreateWorkspace) GetPath() string {
	return filepath.Clean(c.Path)
}
