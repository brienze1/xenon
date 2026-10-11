package entity

import "github.com/brienze1/xenon/internal/domain/enums/pathaction"

type ThreeWayPlan struct {
	Paths []PathPlan
}

func (t ThreeWayPlan) Conflicts() []PathPlan {
	return t.withAction(func(plan PathPlan) bool { return plan.IsConflict() })
}

func (t ThreeWayPlan) StructuralMerges() []PathPlan {
	return t.withAction(func(plan PathPlan) bool { return plan.Action == pathaction.MergeStructural })
}

func (t ThreeWayPlan) withAction(matches func(PathPlan) bool) []PathPlan {
	selected := make([]PathPlan, 0)
	for _, plan := range t.Paths {
		if matches(plan) {
			selected = append(selected, plan)
		}
	}
	return selected
}
