package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/splitorientation"
)

const (
	minSidebarWidth    = 200
	maxSidebarWidth    = 600
	minRightPanelWidth = 240
	maxRightPanelWidth = 600
	minConsoleHeight   = 120
	minSplitRatio      = 0.2
	maxSplitRatio      = 0.8
	infoPanelView      = "info"
)

type Layout struct {
	Split             enums.SplitOrientation
	RightPanelView    string
	SplitRatio        float64
	SidebarWidth      int
	RightPanelWidth   int
	ConsoleHeight     int
	SidebarVisible    bool
	RightPanelVisible bool
	ConsoleVisible    bool
}

func NewLayout() Layout {
	return Layout{
		SidebarWidth:      280,
		SidebarVisible:    true,
		RightPanelWidth:   320,
		RightPanelVisible: true,
		RightPanelView:    infoPanelView,
		ConsoleHeight:     200,
		Split:             splitorientation.Stacked,
		SplitRatio:        0.5,
	}
}

func (l Layout) Normalized() Layout {
	l.SidebarWidth = min(max(l.SidebarWidth, minSidebarWidth), maxSidebarWidth)
	l.RightPanelWidth = min(max(l.RightPanelWidth, minRightPanelWidth), maxRightPanelWidth)
	l.ConsoleHeight = max(l.ConsoleHeight, minConsoleHeight)
	l.SplitRatio = min(max(l.SplitRatio, minSplitRatio), maxSplitRatio)
	l.RightPanelView = infoPanelView
	if l.Split != splitorientation.SideBySide {
		l.Split = splitorientation.Stacked
	}
	return l
}
