package entity

import (
	"slices"
	"sort"
)

type RunnerPreferences struct {
	Config RunConfig
}

func NewRunnerPreferences(config RunConfig) RunnerPreferences {
	return RunnerPreferences{Config: config}
}

// StaleRunnerPreferences lists the remembered targets that are no longer collections or folders of the tree.
func StaleRunnerPreferences(preferences map[string]RunnerPreferences, liveIds []string) []string {
	stale := make([]string, 0)
	for itemId := range preferences {
		if !slices.Contains(liveIds, itemId) {
			stale = append(stale, itemId)
		}
	}
	sort.Strings(stale)
	return stale
}
