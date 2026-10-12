package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncChangeItem struct {
	ItemID   string   `json:"item_id"`
	ItemName string   `json:"item_name"`
	ItemKind string   `json:"item_kind"`
	Change   string   `json:"change"`
	Paths    []string `json:"paths"`
}

func NewSyncChangeItems(items []entity.ChangeItem) []SyncChangeItem {
	converted := make([]SyncChangeItem, 0, len(items))
	for _, item := range items {
		converted = append(converted, SyncChangeItem{
			ItemID:   item.ItemID,
			ItemName: item.ItemName,
			ItemKind: item.ItemKind.String(),
			Change:   item.Change.String(),
			Paths:    append([]string{}, item.Paths...),
		})
	}
	return converted
}
