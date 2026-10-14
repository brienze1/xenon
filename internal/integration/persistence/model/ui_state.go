package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
)

type UIState struct {
	ActiveTabId   string           `json:"active_tab_id"`
	PreviewTabId  string           `json:"preview_tab_id"`
	HistorySource string           `json:"history_source,omitempty"`
	Sidebar       Sidebar          `json:"sidebar"`
	Tabs          []Tab            `json:"tabs"`
	ClosedTabs    []ClosedTab      `json:"closed_tabs"`
	Snippets      []SnippetOptions `json:"snippet_options"`
	Layout        Layout           `json:"layout"`
	Version       int              `json:"version"`
}

func NewUIState(state entity.UIState) UIState {
	model := UIState{
		Version:       state.Version,
		Tabs:          make([]Tab, 0, len(state.Tabs)),
		ActiveTabId:   state.ActiveTabID,
		PreviewTabId:  state.PreviewTabID,
		Sidebar:       NewSidebar(state.Sidebar),
		Layout:        NewLayout(state.Layout),
		HistorySource: state.HistorySource.String(),
	}
	for _, tab := range state.Tabs {
		model.Tabs = append(model.Tabs, NewTab(tab))
	}
	model.ClosedTabs = make([]ClosedTab, 0, len(state.ClosedTabs))
	for _, closed := range state.ClosedTabs {
		model.ClosedTabs = append(model.ClosedTabs, NewClosedTab(closed))
	}
	model.Snippets = make([]SnippetOptions, 0, len(state.Snippets))
	for _, options := range state.Snippets {
		model.Snippets = append(model.Snippets, NewSnippetOptions(options))
	}
	return model
}

func (u *UIState) ToEntity() entity.UIState {
	state := entity.NewUIState()
	state.ActiveTabID = u.ActiveTabId
	state.PreviewTabID = u.PreviewTabId
	for index := range u.Tabs {
		state.Tabs = append(state.Tabs, u.Tabs[index].ToEntity())
	}
	for index := range u.ClosedTabs {
		state.ClosedTabs = append(state.ClosedTabs, u.ClosedTabs[index].ToEntity())
	}
	for index := range u.Snippets {
		state.Snippets = append(state.Snippets, u.Snippets[index].ToEntity())
	}
	state.Sidebar = u.Sidebar.ToEntity()
	state.Layout = u.Layout.ToEntity()
	state.HistorySource = historyfilter.GetHistoryFilter(u.HistorySource)
	return state.Normalized()
}
