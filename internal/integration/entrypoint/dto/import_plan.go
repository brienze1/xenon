package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

type ImportPlanItem struct {
	Key            string `json:"key" validate:"required,not_blank"`
	TargetLocalId  string `json:"target_local_id" validate:"required_if=Include true"`
	TargetParentId string `json:"target_parent_id"`
	Resolution     string `json:"resolution" validate:"omitempty,oneof=replace keep_both skip"`
	Include        bool   `json:"include"`
}

type ImportPlan struct {
	SourceId   string           `json:"source_id" validate:"required,not_blank"`
	ApplyToAll string           `json:"apply_to_all" validate:"omitempty,oneof=replace keep_both skip"`
	Items      []ImportPlanItem `json:"items" validate:"dive"`
}

func (i *ImportPlan) ToEntity() entity.ImportPlan {
	plan := entity.ImportPlan{SourceID: i.SourceId, ApplyToAll: clashresolution.GetClashResolution(i.ApplyToAll), Items: make([]entity.ImportPlanItem, 0, len(i.Items))}
	for _, item := range i.Items {
		plan.Items = append(plan.Items, entity.ImportPlanItem{
			Key:            item.Key,
			TargetLocalID:  item.TargetLocalId,
			TargetParentID: item.TargetParentId,
			Resolution:     clashresolution.GetClashResolution(item.Resolution),
			Include:        item.Include,
		})
	}
	return plan
}

func (i ImportPlan) LogValue() any {
	return map[string]any{"source_id": i.SourceId, "items": len(i.Items), "apply_to_all": i.ApplyToAll}
}
