package entity

import "slices"

const ConsoleCapacity = 5000

type ConsoleBuffer struct {
	Entries []ConsoleEntry
}

func NewConsoleBuffer() ConsoleBuffer {
	return ConsoleBuffer{Entries: []ConsoleEntry{}}
}

// Appended keeps the newest ConsoleCapacity entries.
func (c ConsoleBuffer) Appended(entries []ConsoleEntry) ConsoleBuffer {
	all := append(slices.Clone(c.Entries), entries...)
	return ConsoleBuffer{Entries: all[max(0, len(all)-ConsoleCapacity):]}
}

// After lists up to limit entries following the one with afterId, or from the oldest when that entry is not kept.
func (c ConsoleBuffer) After(afterId string, limit int) []ConsoleEntry {
	start := slices.IndexFunc(c.Entries, func(entry ConsoleEntry) bool { return entry.ID == afterId }) + 1
	end := min(len(c.Entries), start+limit)
	return slices.Clone(c.Entries[start:end])
}
