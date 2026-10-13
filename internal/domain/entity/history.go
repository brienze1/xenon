package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
)

const (
	DefaultHistoryCap      = 1000
	MinHistoryCap          = 100
	MaxHistoryCap          = 20000
	DefaultHistoryPageSize = 1000
	historySlackDivisor    = 5
)

var DefaultHistoryCaps = HistoryCaps{User: DefaultHistoryCap, AI: DefaultHistoryCap}

type HistoryCaps struct {
	User int
	AI   int
}

type History struct {
	entries []HistoryEntry
}

func IsValidHistoryCap(limit int) bool {
	return limit >= MinHistoryCap && limit <= MaxHistoryCap
}

func (h HistoryCaps) OrDefault() HistoryCaps {
	if !IsValidHistoryCap(h.User) {
		h.User = DefaultHistoryCap
	}
	if !IsValidHistoryCap(h.AI) {
		h.AI = DefaultHistoryCap
	}
	return h
}

func (h HistoryCaps) For(source enums.HistorySource) int {
	if source == historysource.AI {
		return h.AI
	}
	return h.User
}

func (h HistoryCaps) compactionThreshold(source enums.HistorySource) int {
	limit := h.For(source)
	return limit + limit/historySlackDivisor
}

// CompactHistory keeps the newest cap entries of each source once one source grows a fifth past its cap.
func CompactHistory[T any](oldestFirst []T, sourceOf func(T) enums.HistorySource, caps HistoryCaps) ([]T, bool) {
	counts := map[enums.HistorySource]int{}
	for _, entry := range oldestFirst {
		counts[sourceOf(entry)]++
	}
	if counts[historysource.User] <= caps.compactionThreshold(historysource.User) && counts[historysource.AI] <= caps.compactionThreshold(historysource.AI) {
		return oldestFirst, false
	}
	return KeepNewestHistory(oldestFirst, sourceOf, caps), true
}

func KeepNewestHistory[T any](oldestFirst []T, sourceOf func(T) enums.HistorySource, caps HistoryCaps) []T {
	kept := make([]T, 0, len(oldestFirst))
	seen := map[enums.HistorySource]int{}
	for index := len(oldestFirst) - 1; index >= 0; index-- {
		source := sourceOf(oldestFirst[index])
		if seen[source] >= caps.For(source) {
			continue
		}
		seen[source]++
		kept = append(kept, oldestFirst[index])
	}
	reversed := make([]T, 0, len(kept))
	for index := len(kept) - 1; index >= 0; index-- {
		reversed = append(reversed, kept[index])
	}
	return reversed
}

func WithoutHistorySource[T any](oldestFirst []T, sourceOf func(T) enums.HistorySource, source enums.HistorySource) []T {
	kept := make([]T, 0, len(oldestFirst))
	for _, entry := range oldestFirst {
		if sourceOf(entry) != source {
			kept = append(kept, entry)
		}
	}
	return kept
}

func HistoryEntrySource(entry HistoryEntry) enums.HistorySource {
	return entry.Origin()
}

func NewHistory(oldestFirst []HistoryEntry, caps HistoryCaps) History {
	kept := KeepNewestHistory(oldestFirst, HistoryEntrySource, caps)
	entries := make([]HistoryEntry, 0, len(kept))
	for index := len(kept) - 1; index >= 0; index-- {
		entries = append(entries, kept[index])
	}
	return History{entries: entries}
}

func (h History) Page(query HistoryQuery) []HistoryEntry {
	page := make([]HistoryEntry, 0, min(query.Limit, len(h.entries)))
	for _, entry := range h.entries[h.startAfter(query):] {
		if query.IsFilledBy(len(page)) {
			break
		}
		if query.Matches(entry) {
			page = append(page, entry)
		}
	}
	return page
}

func (h History) Find(id string) (HistoryEntry, bool) {
	for _, entry := range h.entries {
		if entry.HasID(id) {
			return entry, true
		}
	}
	return HistoryEntry{}, false
}

func (h History) WithoutSource(source enums.HistorySource) []HistoryEntry {
	return WithoutHistorySource(h.entries, HistoryEntrySource, source)
}

func (h History) startAfter(query HistoryQuery) int {
	if !query.HasCursor() {
		return 0
	}
	for index, entry := range h.entries {
		if query.IsCursor(entry) {
			return index + 1
		}
	}
	return len(h.entries)
}
