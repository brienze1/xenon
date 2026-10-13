package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/tabkind"
)

type Tab struct {
	Draft         *Request `json:"draft,omitempty"`
	TabId         string   `json:"tab_id" validate:"required,not_blank"`
	Kind          string   `json:"kind"`
	RequestId     string   `json:"request_id"`
	ExampleId     string   `json:"example_id"`
	EnvironmentId string   `json:"environment_id"`
	ItemId        string   `json:"item_id"`
	BaseRevision  string   `json:"base_revision"`
	Pinned        bool     `json:"pinned"`
}

func NewTab(tab entity.Tab) Tab {
	result := Tab{TabId: tab.TabID, Kind: entity.NewTabKind(tab.Kind, tab.ExampleID, tab.EnvironmentID, tab.ItemID).String(), RequestId: tab.RequestID, ExampleId: tab.ExampleID, EnvironmentId: tab.EnvironmentID, ItemId: tab.ItemID, BaseRevision: tab.BaseRevision, Pinned: tab.Pinned}
	if tab.Draft != nil {
		draft := NewRequest(*tab.Draft)
		result.Draft = &draft
	}
	return result
}

func (t *Tab) ToEntity() entity.Tab {
	tab := entity.Tab{TabID: t.TabId, Kind: entity.NewTabKind(tabkind.GetTabKind(t.Kind), t.ExampleId, t.EnvironmentId, t.ItemId), RequestID: t.RequestId, ExampleID: t.ExampleId, EnvironmentID: t.EnvironmentId, ItemID: t.ItemId, BaseRevision: t.BaseRevision, Pinned: t.Pinned}
	if t.Draft != nil {
		draft := t.Draft.ToEntity()
		tab.Draft = &draft
	}
	return tab
}
