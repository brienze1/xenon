package model

import "github.com/brienze1/xenon/internal/domain/entity"

type Tab struct {
	Draft        *Request `json:"draft,omitempty"`
	TabId        string   `json:"tab_id"`
	Kind         string   `json:"kind"`
	RequestId    string   `json:"request_id"`
	ExampleId    string   `json:"example_id"`
	BaseRevision string   `json:"base_revision"`
	Pinned       bool     `json:"pinned"`
}

func NewTab(tab entity.Tab) Tab {
	model := Tab{TabId: tab.TabID, Kind: entity.NewTabKind(tab.ExampleID).String(), RequestId: tab.RequestID, ExampleId: tab.ExampleID, BaseRevision: tab.BaseRevision, Pinned: tab.Pinned}
	if tab.Draft != nil {
		draft := NewRequest(*tab.Draft)
		model.Draft = &draft
	}
	return model
}

func (t *Tab) ToEntity() entity.Tab {
	tab := entity.Tab{TabID: t.TabId, Kind: entity.NewTabKind(t.ExampleId), RequestID: t.RequestId, ExampleID: t.ExampleId, BaseRevision: t.BaseRevision, Pinned: t.Pinned}
	if t.Draft != nil {
		draft := t.Draft.ToEntity()
		tab.Draft = &draft
	}
	return tab
}
