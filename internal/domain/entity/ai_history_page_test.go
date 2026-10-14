package entity

import (
	"strconv"
	"testing"
)

func historyEntries(count int) []HistoryEntry {
	entries := make([]HistoryEntry, 0, count)
	for index := range count {
		entries = append(entries, HistoryEntry{ID: strconv.Itoa(index)})
	}
	return entries
}

func TestNewAIHistoryPage(t *testing.T) {
	cases := []struct {
		fetched int
		items   int
		next    string
	}{
		{fetched: 0, items: 0},
		{fetched: 5, items: 5},
		{fetched: AIPageSize, items: AIPageSize},
		{fetched: AIPageSize + 1, items: AIPageSize, next: "99"},
	}
	for _, tc := range cases {
		page := NewAIHistoryPage(historyEntries(tc.fetched))
		if len(page.Items) != tc.items || page.Next != tc.next || page.Items == nil {
			t.Errorf("%d fetched: %d items, next %q", tc.fetched, len(page.Items), page.Next)
		}
	}
}

func TestAppRegistryActive(t *testing.T) {
	registry := NewAppRegistry()
	if _, found := registry.Active(); found {
		t.Error("an empty registry has no active workspace")
	}
	registry.Register(RegisteredWorkspace{LocalID: "w1", DisplayNameCache: "One"})
	registry.Register(RegisteredWorkspace{LocalID: "w2", DisplayNameCache: "Two"})
	if active, found := registry.Active(); !found || active.LocalID != "w2" || active.DisplayNameCache != "Two" {
		t.Errorf("Active() = %+v %v", active, found)
	}
	registry.ActiveLocalID = "gone"
	if _, found := registry.Active(); found {
		t.Error("an unregistered active id is not active")
	}
}
