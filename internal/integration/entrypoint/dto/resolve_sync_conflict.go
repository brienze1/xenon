package dto

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
)

type ResolveSyncConflict struct {
	SessionId  string `json:"session_id" validate:"required,not_blank"`
	ItemId     string `json:"item_id" validate:"required,not_blank"`
	Resolution string `json:"resolution" validate:"required,oneof=mine theirs both keep delete"`
	FieldPath  string `json:"field_path"`
}

func (r *ResolveSyncConflict) ConflictResolution() enums.ConflictResolution {
	return conflictresolution.GetConflictResolution(r.Resolution)
}
