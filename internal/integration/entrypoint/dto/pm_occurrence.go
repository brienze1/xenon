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

func NewPmOccurrence(occurrence entity.PmOccurrence) PmOccurrence {
	return PmOccurrence{
		ItemKey: occurrence.ItemKey,
		ItemId:  occurrence.ItemID,
		LocalId: occurrence.LocalID,
		Path:    occurrence.Path,
		Phase:   occurrence.Phase.String(),
		Line:    occurrence.Line,
	}
}
