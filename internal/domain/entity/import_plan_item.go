package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ImportPlanItem struct {
	Key            string
	TargetLocalID  string
	TargetParentID string
	TargetFolder   string
	Name           string
	Resolution     enums.ClashResolution
	Include        bool
}

func (i ImportPlanItem) HasTargetFolder() bool {
	return i.TargetFolder != ""
}

func (i ImportPlanItem) IsIntoParent() bool {
	return i.TargetParentID != ""
}
