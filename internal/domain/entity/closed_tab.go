package entity

import "time"

const maxClosedTabs = 20

type ClosedTab struct {
	ClosedAt time.Time
	Tab      Tab
}

func LatestClosedTabs(tabs []ClosedTab) []ClosedTab {
	if len(tabs) <= maxClosedTabs {
		return tabs
	}
	return tabs[len(tabs)-maxClosedTabs:]
}
