package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

type ImportPlanItem struct {
	Key            string `json:"key" validate:"required,not_blank"`
	TargetLocalId  string `json:"target_local_id" validate:"required_if=Include true"`
	TargetParentId string `json:"target_parent_id"`
	Name           string `json:"name" validate:"max=200"`
	Resolution     string `json:"resolution" validate:"omitempty,oneof=replace keep_both skip"`
	Include        bool   `json:"include"`
}

func (i *ImportPlanItem) ToEntity() entity.ImportPlanItem {
	return entity.ImportPlanItem{
		Key:            i.Key,
		TargetLocalID:  i.TargetLocalId,
		TargetParentID: i.TargetParentId,
		Name:           i.Name,
		Resolution:     clashresolution.GetClashResolution(i.Resolution),
		Include:        i.Include,
	}
}
