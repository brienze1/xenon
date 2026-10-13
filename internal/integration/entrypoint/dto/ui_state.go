package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type UIState struct {
	LocalId      string           `json:"local_id" validate:"required,not_blank"`
	ActiveTabId  string           `json:"active_tab_id"`
	PreviewTabId string           `json:"preview_tab_id"`
	Tabs         []Tab            `json:"tabs" validate:"dive"`
	ClosedTabs   []ClosedTab      `json:"closed_tabs" validate:"dive"`
	Snippets     []SnippetOptions `json:"snippet_options" validate:"dive"`
	Sidebar      SidebarState     `json:"sidebar"`
	Layout       Layout           `json:"layout"`
	Version      int              `json:"version"`
}

func NewUIState(localId string, state entity.UIState) UIState {
	result := UIState{
		LocalId:      localId,
		Version:      state.Version,
		ActiveTabId:  state.ActiveTabID,
		PreviewTabId: state.PreviewTabID,
		Tabs:         make([]Tab, 0, len(state.Tabs)),
		Sidebar:      NewSidebarState(state.Sidebar),
		Layout:       NewLayout(state.Layout),
	}
	for _, tab := range state.Tabs {
		result.Tabs = append(result.Tabs, NewTab(tab))
	}
	result.ClosedTabs = make([]ClosedTab, 0, len(state.ClosedTabs))
	for _, closed := range state.ClosedTabs {
		result.ClosedTabs = append(result.ClosedTabs, NewClosedTab(closed))
	}
	result.Snippets = make([]SnippetOptions, 0, len(state.Snippets))
	for _, options := range state.Snippets {
		result.Snippets = append(result.Snippets, NewSnippetOptions(options))
	}
	return result
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
	return state
}

func (u UIState) LogValue() any {
	return map[string]any{"local_id": u.LocalId, "tabs": len(u.Tabs), "active_tab_id": u.ActiveTabId}
}
