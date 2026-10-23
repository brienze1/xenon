package entity

import (
	"slices"
	"strings"
)

const (
	deletedEnvironmentLabel = "Deleted environment"
	aliasEnvironmentsNote   = "Xenon alias for environments: "
)

// NormalizedEnvironmentIDs trims the ids and drops blanks and repeats, keeping the first order; none means all environments.
func NormalizedEnvironmentIDs(ids []string) []string {
	var normalized []string
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" && !slices.Contains(normalized, trimmed) {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

// AliasEnvironmentLabels names each id by its environment, or as a deleted environment when none has it.
func AliasEnvironmentLabels(ids []string, environments []Environment) []string {
	labels := make([]string, 0, len(ids))
	for _, id := range ids {
		index := slices.IndexFunc(environments, func(environment Environment) bool { return environment.ID == id })
		if index < 0 {
			labels = append(labels, deletedEnvironmentLabel)
			continue
		}
		labels = append(labels, environments[index].Name)
	}
	return labels
}
