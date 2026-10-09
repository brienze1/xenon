package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/splitorientation"
)

type Layout struct {
	Split             string  `json:"split"`
	RightPanelView    string  `json:"right_panel_view"`
	SplitRatio        float64 `json:"split_ratio"`
	SidebarWidth      int     `json:"sidebar_width"`
	RightPanelWidth   int     `json:"right_panel_width"`
	ConsoleHeight     int     `json:"console_height"`
	SidebarVisible    bool    `json:"sidebar_visible"`
	RightPanelVisible bool    `json:"right_panel_visible"`
	ConsoleVisible    bool    `json:"console_visible"`
}

func NewLayout(layout entity.Layout) Layout {
	return Layout{
		Split:             layout.Split.String(),
		RightPanelView:    layout.RightPanelView,
		SplitRatio:        layout.SplitRatio,
		SidebarWidth:      layout.SidebarWidth,
		RightPanelWidth:   layout.RightPanelWidth,
		ConsoleHeight:     layout.ConsoleHeight,
		SidebarVisible:    layout.SidebarVisible,
		RightPanelVisible: layout.RightPanelVisible,
		ConsoleVisible:    layout.ConsoleVisible,
	}
}

func (l *Layout) ToEntity() entity.Layout {
	return entity.Layout{
		Split:             splitorientation.GetSplitOrientation(l.Split),
		RightPanelView:    l.RightPanelView,
		SplitRatio:        l.SplitRatio,
		SidebarWidth:      l.SidebarWidth,
		RightPanelWidth:   l.RightPanelWidth,
		ConsoleHeight:     l.ConsoleHeight,
		SidebarVisible:    l.SidebarVisible,
		RightPanelVisible: l.RightPanelVisible,
		ConsoleVisible:    l.ConsoleVisible,
	}
}
