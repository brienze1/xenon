package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPGitConflictField struct {
	Path       string `json:"path"`
	Resolution string `json:"resolution,omitempty"`
}

func NewMCPGitConflictField(field entity.FieldConflict) MCPGitConflictField {
	return MCPGitConflictField{Path: field.Path, Resolution: field.ChosenResolution().String()}
}
