package entity

import "slices"

const ConsoleCapacity = 5000

type ConsoleBuffer struct {
	Entries []ConsoleEntry
}

func NewConsoleBuffer() ConsoleBuffer {
	return ConsoleBuffer{Entries: []ConsoleEntry{}}
}

// Appended keeps the newest ConsoleCapacity entries and hands back the ones it dropped.
func (c ConsoleBuffer) Appended(entries []ConsoleEntry) (ConsoleBuffer, []ConsoleEntry) {
	all := append(slices.Clone(c.Entries), entries...)
	cut := max(0, len(all)-ConsoleCapacity)
	return ConsoleBuffer{Entries: all[cut:]}, all[:cut]
}

// Find lists up to query.Limit matching entries following the one with query.AfterID, or from the oldest when that entry is not kept.
func (c ConsoleBuffer) Find(query ConsoleQuery) []ConsoleEntry {
	start := slices.IndexFunc(c.Entries, func(entry ConsoleEntry) bool { return entry.ID == query.AfterID }) + 1
	found := make([]ConsoleEntry, 0, min(query.Limit, len(c.Entries)-start))
	for _, entry := range c.Entries[start:] {
		if len(found) == query.Limit {
			break
		}
		if query.Matches(entry) {
			found = append(found, entry)
		}
	}
	return found
}

func (c ConsoleBuffer) Entry(id string) (ConsoleEntry, bool) {
	index := slices.IndexFunc(c.Entries, func(entry ConsoleEntry) bool { return entry.ID == id })
	if index < 0 {
		return ConsoleEntry{}, false
	}
	return c.Entries[index], true
}
