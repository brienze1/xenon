package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type Tab struct {
	Draft        *Request
	Kind         enums.ItemKind
	TabID        string
	RequestID    string
	ExampleID    string
	BaseRevision string
	Pinned       bool
}

func NewTabKind(exampleId string) enums.ItemKind {
	if exampleId != "" {
		return itemkind.Example
	}
	return itemkind.Request
}

func (t Tab) IsPinned() bool {
	return t.Pinned
}

func (t Tab) HasID(id string) bool {
	return t.TabID == id
}

func (t Tab) LogValue() any {
	value := map[string]any{"tabId": t.TabID, "kind": t.Kind, "requestId": t.RequestID, "exampleId": t.ExampleID, "baseRevision": t.BaseRevision, "pinned": t.Pinned}
	if t.Draft != nil {
		value["draft"] = t.Draft.LogValue()
	}
	return value
}
