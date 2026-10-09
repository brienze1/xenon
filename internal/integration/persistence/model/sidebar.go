package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/sidebarview"
)

type Sidebar struct {
	View       string   `json:"view"`
	SelectedId string   `json:"selected_id"`
	Search     string   `json:"search"`
	Expanded   []string `json:"expanded"`
}

func NewSidebar(sidebar entity.SidebarState) Sidebar {
	return Sidebar{View: sidebar.View.String(), SelectedId: sidebar.SelectedID, Search: sidebar.Search, Expanded: nonNil(sidebar.Expanded)}
}

func (s *Sidebar) ToEntity() entity.SidebarState {
	return entity.SidebarState{
		View:       sidebarview.GetSidebarView(s.View),
		SelectedID: s.SelectedId,
		Search:     s.Search,
		Expanded:   nonNil(s.Expanded),
	}
}
