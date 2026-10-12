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
	BaseRevision  string
	Pinned        bool
}

func NewTabKind(kind enums.TabKind, exampleId, environmentId string) enums.TabKind {
	switch {
	case kind == tabkind.Globals:
		return tabkind.Globals
	case environmentId != "":
		return tabkind.Environment
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
	value := map[string]any{"tabId": t.TabID, "kind": t.Kind, "requestId": t.RequestID, "exampleId": t.ExampleID, "environmentId": t.EnvironmentID, "baseRevision": t.BaseRevision, "pinned": t.Pinned}
	if t.Draft != nil {
		value["draft"] = t.Draft.LogValue()
	}
	return value
}
