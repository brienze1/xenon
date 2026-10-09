package entity

const uiStateVersion = 1

type UIState struct {
	ActiveTabID  string
	PreviewTabID string
	Tabs         []Tab
	Sidebar      SidebarState
	Layout       Layout
	Version      int
}

func NewUIState() UIState {
	return UIState{Version: uiStateVersion, Tabs: []Tab{}, Sidebar: NewSidebarState(), Layout: NewLayout()}
}

func (u UIState) Normalized() UIState {
	u.Version = uiStateVersion
	u.Tabs = uniqueTabs(u.Tabs)
	if _, found := u.tab(u.ActiveTabID); !found {
		u.ActiveTabID = u.firstTabID()
	}
	if preview, found := u.tab(u.PreviewTabID); !found || preview.IsPinned() {
		u.PreviewTabID = ""
	}
	u.Sidebar = u.Sidebar.Normalized()
	u.Layout = u.Layout.Normalized()
	return u
}

func (u UIState) tab(id string) (Tab, bool) {
	for _, tab := range u.Tabs {
		if tab.HasID(id) {
			return tab, true
		}
	}
	return Tab{}, false
}

func (u UIState) firstTabID() string {
	if len(u.Tabs) == 0 {
		return ""
	}
	return u.Tabs[0].TabID
}

func uniqueTabs(tabs []Tab) []Tab {
	unique := make([]Tab, 0, len(tabs))
	seen := map[string]bool{}
	for _, tab := range tabs {
		if seen[tab.TabID] {
			continue
		}
		seen[tab.TabID] = true
		unique = append(unique, tab)
	}
	return unique
}

func (u UIState) LogValue() any {
	tabs := make([]any, 0, len(u.Tabs))
	for _, tab := range u.Tabs {
		tabs = append(tabs, tab.LogValue())
	}
	return map[string]any{
		"activeTabId":  u.ActiveTabID,
		"previewTabId": u.PreviewTabID,
		"tabs":         tabs,
		"sidebar":      u.Sidebar,
		"layout":       u.Layout,
		"version":      u.Version,
	}
}
