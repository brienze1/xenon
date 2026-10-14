package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/tabkind"
)

type Tab struct {
	Draft         *Request
	Kind          enums.TabKind
	TabID         string
	RequestID     string
	ExampleID     string
	EnvironmentID string
	ItemID        string
	BaseRevision  string
	Pinned        bool
}

func NewTabKind(kind enums.TabKind, exampleId, environmentId, itemId string) enums.TabKind {
	switch {
	case kind == tabkind.Globals:
		return tabkind.Globals
	case kind == tabkind.Settings:
		return tabkind.Settings
	case environmentId != "":
		return tabkind.Environment
	case itemId != "" && kind == tabkind.Folder:
		return tabkind.Folder
	case itemId != "":
		return tabkind.Collection
	case exampleId != "":
		return tabkind.Example
	default:
		return tabkind.Request
	}
}

func (t Tab) IsPinned() bool {
	return t.Pinned
}

func (t Tab) HasID(id string) bool {
	return t.TabID == id
}

func (t Tab) LogValue() any {
	value := map[string]any{"tabId": t.TabID, "kind": t.Kind, "requestId": t.RequestID, "exampleId": t.ExampleID, "environmentId": t.EnvironmentID, "itemId": t.ItemID, "baseRevision": t.BaseRevision, "pinned": t.Pinned}
	if t.Draft != nil {
		value["draft"] = t.Draft.LogValue()
	}
	return value
}
