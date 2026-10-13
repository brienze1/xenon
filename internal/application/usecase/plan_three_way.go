package usecase

import (
	"sort"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func PlanThreeWay(base, mine, theirs map[string]string) entity.ThreeWayPlan {
	paths := map[string]bool{}
	for _, versions := range []map[string]string{base, mine, theirs} {
		for path := range versions {
			paths[path] = true
		}
	}
	sorted := make([]string, 0, len(paths))
	for path := range paths {
		sorted = append(sorted, path)
	}
	sort.Strings(sorted)
	plan := entity.ThreeWayPlan{Paths: make([]entity.PathPlan, 0, len(sorted))}
	for _, path := range sorted {
		plan.Paths = append(plan.Paths, entity.NewPathPlan(path, base[path], mine[path], theirs[path]))
	}
	return plan
}
