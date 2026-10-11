package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

type ImportPlan struct {
	SourceId    string           `json:"source_id" validate:"required,not_blank"`
	ApplyToAll  string           `json:"apply_to_all" validate:"omitempty,oneof=replace keep_both skip"`
	Passphrase  string           `json:"passphrase"`
	Items       []ImportPlanItem `json:"items" validate:"dive"`
	SkipSecrets bool             `json:"skip_secrets"`
}

func (i *ImportPlan) ToEntity() entity.ImportPlan {
	plan := entity.ImportPlan{
		SourceID:    i.SourceId,
		ApplyToAll:  clashresolution.GetClashResolution(i.ApplyToAll),
		Passphrase:  i.Passphrase,
		SkipSecrets: i.SkipSecrets,
		Items:       make([]entity.ImportPlanItem, 0, len(i.Items)),
	}
	for index := range i.Items {
		plan.Items = append(plan.Items, i.Items[index].ToEntity())
	}
	return plan
}

func (i ImportPlan) LogValue() any {
	return map[string]any{"source_id": i.SourceId, "items": len(i.Items), "apply_to_all": i.ApplyToAll, "skip_secrets": i.SkipSecrets}
}
