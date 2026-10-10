package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportClash struct {
	ItemKey      string `json:"item_key"`
	Kind         string `json:"kind"`
	ExistingId   string `json:"existing_id"`
	ExistingName string `json:"existing_name"`
	Name         string `json:"name"`
}

type ImportClashes struct {
	Clashes []ImportClash `json:"clashes"`
}

func NewImportClashes(clashes []entity.Clash) ImportClashes {
	result := ImportClashes{Clashes: make([]ImportClash, 0, len(clashes))}
	for _, clash := range clashes {
		result.Clashes = append(result.Clashes, ImportClash{
			ItemKey:      clash.ItemKey,
			Kind:         clash.Kind.String(),
			ExistingId:   clash.ExistingID,
			ExistingName: clash.ExistingName,
			Name:         clash.Name,
		})
	}
	return result
}
