package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type PmOccurrence struct {
	ItemKey string `json:"item_key"`
	ItemId  string `json:"item_id"`
	LocalId string `json:"local_id"`
	Path    string `json:"path"`
	Phase   string `json:"phase"`
	Line    int    `json:"line"`
}

type PmUsage struct {
	Api         string         `json:"api"`
	Occurrences []PmOccurrence `json:"occurrences"`
}

func NewPmUsages(usages []entity.PmUsage) []PmUsage {
	result := make([]PmUsage, 0, len(usages))
	for _, usage := range usages {
		occurrences := make([]PmOccurrence, 0, len(usage.Occurrences))
		for _, occurrence := range usage.Occurrences {
			occurrences = append(occurrences, PmOccurrence{
				ItemKey: occurrence.ItemKey,
				ItemId:  occurrence.ItemID,
				LocalId: occurrence.LocalID,
				Path:    occurrence.Path,
				Phase:   occurrence.Phase.String(),
				Line:    occurrence.Line,
			})
		}
		result = append(result, PmUsage{Api: usage.API, Occurrences: occurrences})
	}
	return result
}
