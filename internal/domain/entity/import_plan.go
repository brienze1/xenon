package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

type ImportPlanItem struct {
	Key            string
	TargetLocalID  string
	TargetParentID string
	Name           string
	Resolution     enums.ClashResolution
	Include        bool
}

func (i ImportPlanItem) IsIntoParent() bool {
	return i.TargetParentID != ""
}

type ImportPlan struct {
	SourceID   string
	ApplyToAll enums.ClashResolution
	Items      []ImportPlanItem
}

func (i ImportPlan) Included() []ImportPlanItem {
	included := make([]ImportPlanItem, 0, len(i.Items))
	for _, item := range i.Items {
		if item.Include {
			included = append(included, item)
		}
	}
	return included
}

func (i ImportPlan) ResolutionFor(item ImportPlanItem) enums.ClashResolution {
	if item.Resolution != clashresolution.Unresolved {
		return item.Resolution
	}
	return i.ApplyToAll
}

func (i ImportPlan) Resolves(item ImportPlanItem) bool {
	return i.ResolutionFor(item) != clashresolution.Unresolved
}

func (i ImportPlan) LogValue() any {
	return map[string]any{"sourceId": i.SourceID, "items": len(i.Items), "applyToAll": i.ApplyToAll}
}
