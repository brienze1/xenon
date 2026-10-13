package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportClashes struct {
	Clashes []ImportClash `json:"clashes"`
}

func NewImportClashes(clashes []entity.Clash) ImportClashes {
	result := ImportClashes{Clashes: make([]ImportClash, 0, len(clashes))}
	for _, clash := range clashes {
		result.Clashes = append(result.Clashes, NewImportClash(clash))
	}
	return result
}
