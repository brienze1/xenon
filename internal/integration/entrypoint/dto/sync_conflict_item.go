package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncConflictItem struct {
	ItemID     string `json:"item_id"`
	ItemName   string `json:"item_name"`
	ItemKind   string `json:"item_kind"`
	Kind       string `json:"kind"`
	Resolution string `json:"resolution"`
	Path       string `json:"path"`
}

func NewSyncConflictItem(item entity.ConflictItem) SyncConflictItem {
	return SyncConflictItem{
		ItemID:     item.ItemID,
		ItemName:   item.ItemName,
		ItemKind:   item.ItemKind.String(),
		Kind:       item.Kind.String(),
		Resolution: item.Resolution.String(),
		Path:       item.Path,
	}
}
