package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncFieldConflict struct {
	Mine       *string `json:"mine"`
	Theirs     *string `json:"theirs"`
	Path       string  `json:"path"`
	Resolution string  `json:"resolution"`
}

func NewSyncFieldConflict(field entity.FieldConflict) SyncFieldConflict {
	return SyncFieldConflict{Mine: field.MineText, Theirs: field.TheirsText, Path: field.Path, Resolution: field.Resolution.String()}
}
