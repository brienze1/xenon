package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/sidebarview"
)

type SidebarState struct {
	View       string   `json:"view" validate:"omitempty,oneof=collections history environments"`
	SelectedId string   `json:"selected_id"`
	Search     string   `json:"search"`
	Expanded   []string `json:"expanded"`
}

func NewSidebarState(sidebar entity.SidebarState) SidebarState {
	return SidebarState{View: sidebar.View.String(), SelectedId: sidebar.SelectedID, Search: sidebar.Search, Expanded: sidebar.Expanded}
}

func (s *SidebarState) ToEntity() entity.SidebarState {
	return entity.SidebarState{
		View:       sidebarview.GetSidebarView(s.View),
		SelectedID: s.SelectedId,
		Search:     s.Search,
		Expanded:   append([]string{}, s.Expanded...),
	}
}
