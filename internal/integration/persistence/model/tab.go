package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/tabkind"
)

type Tab struct {
	Draft         *Request           `json:"draft,omitempty"`
	VersionDrafts map[string]Request `json:"version_drafts,omitempty"`
	TabId         string             `json:"tab_id"`
	Kind          string             `json:"kind"`
	RequestId     string             `json:"request_id"`
	ExampleId     string             `json:"example_id"`
	EnvironmentId string             `json:"environment_id"`
	ItemId        string             `json:"item_id"`
	BaseRevision  string             `json:"base_revision"`
	AliasId       string             `json:"alias_id,omitempty"`
	Pinned        bool               `json:"pinned"`
}

func NewTab(tab entity.Tab) Tab {
	model := Tab{TabId: tab.TabID, Kind: entity.NewTabKind(tab.Kind, tab.ExampleID, tab.EnvironmentID, tab.ItemID).String(), RequestId: tab.RequestID, ExampleId: tab.ExampleID, EnvironmentId: tab.EnvironmentID, ItemId: tab.ItemID, BaseRevision: tab.BaseRevision, AliasId: tab.AliasID, VersionDrafts: newVersionDrafts(tab.VersionDrafts), Pinned: tab.Pinned}
	if tab.Draft != nil {
		draft := NewRequest(*tab.Draft)
		model.Draft = &draft
	}
	return model
}

func (t *Tab) ToEntity() entity.Tab {
	tab := entity.Tab{TabID: t.TabId, Kind: entity.NewTabKind(tabkind.GetTabKind(t.Kind), t.ExampleId, t.EnvironmentId, t.ItemId), RequestID: t.RequestId, ExampleID: t.ExampleId, EnvironmentID: t.EnvironmentId, ItemID: t.ItemId, BaseRevision: t.BaseRevision, AliasID: t.AliasId, VersionDrafts: versionDraftsToEntity(t.VersionDrafts), Pinned: t.Pinned}
	if t.Draft != nil {
		draft := t.Draft.ToEntity()
		tab.Draft = &draft
	}
	return tab
}

func newVersionDrafts(drafts map[string]entity.Request) map[string]Request {
	if len(drafts) == 0 {
		return nil
	}
	result := make(map[string]Request, len(drafts))
	for aliasId, draft := range drafts {
		result[aliasId] = NewRequest(draft)
	}
	return result
}

func versionDraftsToEntity(drafts map[string]Request) map[string]entity.Request {
	if len(drafts) == 0 {
		return nil
	}
	result := make(map[string]entity.Request, len(drafts))
	for aliasId, draft := range drafts {
		result[aliasId] = draft.ToEntity()
	}
	return result
}
