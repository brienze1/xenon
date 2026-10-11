package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/rightpanelview"
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
)

type Layout struct {
	Split             enums.SplitOrientation
	RightPanelView    enums.RightPanelView
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
		RightPanelView:    rightpanelview.Info,
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
	if l.RightPanelView != rightpanelview.Variables && l.RightPanelView != rightpanelview.Code {
		l.RightPanelView = rightpanelview.Info
	}
	if l.Split != splitorientation.SideBySide {
		l.Split = splitorientation.Stacked
	}
	return l
}
