package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type PmUsage struct {
	Api         string         `json:"api"`
	Occurrences []PmOccurrence `json:"occurrences"`
}

func NewPmUsage(usage entity.PmUsage) PmUsage {
	occurrences := make([]PmOccurrence, 0, len(usage.Occurrences))
	for _, occurrence := range usage.Occurrences {
		occurrences = append(occurrences, NewPmOccurrence(occurrence))
	}
	return PmUsage{Api: usage.API, Occurrences: occurrences}
}

func NewPmUsages(usages []entity.PmUsage) []PmUsage {
	result := make([]PmUsage, 0, len(usages))
	for _, usage := range usages {
		result = append(result, NewPmUsage(usage))
	}
	return result
}
