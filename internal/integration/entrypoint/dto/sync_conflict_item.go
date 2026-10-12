package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SyncConflictItem struct {
	Base        *SyncItemVersion    `json:"base,omitempty"`
	Mine        *SyncItemVersion    `json:"mine,omitempty"`
	Theirs      *SyncItemVersion    `json:"theirs,omitempty"`
	ItemID      string              `json:"item_id"`
	ItemName    string              `json:"item_name"`
	ItemKind    string              `json:"item_kind"`
	Kind        string              `json:"kind"`
	Resolution  string              `json:"resolution"`
	Path        string              `json:"path"`
	Fields      []SyncFieldConflict `json:"fields"`
	Uncommitted bool                `json:"uncommitted"`
}

func NewSyncConflictItem(item entity.ConflictItem) SyncConflictItem {
	fields := make([]SyncFieldConflict, 0, len(item.Fields))
	for _, field := range item.Fields {
		fields = append(fields, NewSyncFieldConflict(field))
	}
	return SyncConflictItem{
		Base:        NewSyncItemVersion(item.Versions.Base),
		Mine:        NewSyncItemVersion(item.Versions.Mine),
		Theirs:      NewSyncItemVersion(item.Versions.Theirs),
		ItemID:      item.ItemID,
		ItemName:    item.ItemName,
		ItemKind:    item.ItemKind.String(),
		Kind:        item.Kind.String(),
		Resolution:  item.Resolution.String(),
		Path:        item.Path,
		Fields:      fields,
		Uncommitted: item.Uncommitted,
	}
}

func (s SyncConflictItem) LogValue() any {
	return map[string]any{"itemId": s.ItemID, "kind": s.Kind, "resolution": s.Resolution}
}
