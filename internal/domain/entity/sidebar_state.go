package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/sidebarview"
)

type SidebarState struct {
	View       enums.SidebarView
	SelectedID string
	Search     string
	Expanded   []string
}

func NewSidebarState() SidebarState {
	return SidebarState{View: sidebarview.Collections, Expanded: []string{}}
}

func (s SidebarState) Normalized() SidebarState {
	if s.View != sidebarview.History && s.View != sidebarview.Environments {
		s.View = sidebarview.Collections
	}
	if s.Expanded == nil {
		s.Expanded = []string{}
	}
	return s
}
