package entity

import "github.com/brienze1/xenon/internal/domain/enums"

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
