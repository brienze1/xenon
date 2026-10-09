package model

import "github.com/brienze1/xenon/internal/domain/entity"

type Tab struct {
	Draft        *Request `json:"draft,omitempty"`
	TabId        string   `json:"tab_id"`
	RequestId    string   `json:"request_id"`
	BaseRevision string   `json:"base_revision"`
	Pinned       bool     `json:"pinned"`
}

func NewTab(tab entity.Tab) Tab {
	model := Tab{TabId: tab.TabID, RequestId: tab.RequestID, BaseRevision: tab.BaseRevision, Pinned: tab.Pinned}
	if tab.Draft != nil {
		draft := NewRequest(*tab.Draft)
		model.Draft = &draft
	}
	return model
}

func (t *Tab) ToEntity() entity.Tab {
	tab := entity.Tab{TabID: t.TabId, RequestID: t.RequestId, BaseRevision: t.BaseRevision, Pinned: t.Pinned}
	if t.Draft != nil {
		draft := t.Draft.ToEntity()
		tab.Draft = &draft
	}
	return tab
}
