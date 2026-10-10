package entity

import (
	"slices"
	"sort"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type PmOccurrence struct {
	Phase   enums.ScriptPhase
	ItemKey string
	ItemID  string
	LocalID string
	Path    string
	Line    int
}

type PmUsage struct {
	API         string
	Occurrences []PmOccurrence
}

func (p PmUsage) ForItem(key string) PmUsage {
	p.Occurrences = mapOccurrences(p.Occurrences, func(occurrence PmOccurrence) PmOccurrence {
		occurrence.ItemKey = key
		return occurrence
	})
	return p
}

func (p PmUsage) InWorkspace(localId string) PmUsage {
	p.Occurrences = mapOccurrences(p.Occurrences, func(occurrence PmOccurrence) PmOccurrence {
		occurrence.LocalID = localId
		return occurrence
	})
	return p
}

func (p PmUsage) Relocated(move ImportRelocation) PmUsage {
	p.Occurrences = mapOccurrences(p.Occurrences, func(occurrence PmOccurrence) PmOccurrence {
		occurrence.ItemID, occurrence.Path = move.Apply(occurrence.ItemID, occurrence.Path)
		return occurrence
	})
	return p
}

func (p PmUsage) Within(ids []string) PmUsage {
	kept := make([]PmOccurrence, 0, len(p.Occurrences))
	for _, occurrence := range p.Occurrences {
		if slices.Contains(ids, occurrence.ItemID) {
			kept = append(kept, occurrence)
		}
	}
	p.Occurrences = kept
	return p
}

func MergePmUsages(usages []PmUsage) []PmUsage {
	merged := make([]PmUsage, 0, len(usages))
	positions := map[string]int{}
	for _, usage := range usages {
		position, found := positions[usage.API]
		if !found {
			positions[usage.API] = len(merged)
			merged = append(merged, PmUsage{API: usage.API, Occurrences: []PmOccurrence{}})
			position = len(merged) - 1
		}
		merged[position].Occurrences = append(merged[position].Occurrences, usage.Occurrences...)
	}
	sort.SliceStable(merged, func(left, right int) bool { return merged[left].API < merged[right].API })
	return merged
}

func mapOccurrences(occurrences []PmOccurrence, transform func(PmOccurrence) PmOccurrence) []PmOccurrence {
	mapped := make([]PmOccurrence, 0, len(occurrences))
	for _, occurrence := range occurrences {
		mapped = append(mapped, transform(occurrence))
	}
	return mapped
}
