package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportClash struct {
	ItemKey      string `json:"item_key"`
	Kind         string `json:"kind"`
	ExistingId   string `json:"existing_id"`
	ExistingName string `json:"existing_name"`
	Name         string `json:"name"`
}

func NewImportClash(clash entity.Clash) ImportClash {
	return ImportClash{
		ItemKey:      clash.ItemKey,
		Kind:         clash.Kind.String(),
		ExistingId:   clash.ExistingID,
		ExistingName: clash.ExistingName,
		Name:         clash.Name,
	}
}
