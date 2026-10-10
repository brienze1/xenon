package dto

type BruBlock struct {
	Name     string
	Text     string
	Entries  []BruEntry
	Line     int
	TextLine int
}

func NewBruBlock(name string, line int) BruBlock {
	return BruBlock{Name: name, Line: line, Entries: []BruEntry{}}
}

func NewTextBruBlock(name string, line int, text string) BruBlock {
	return BruBlock{Name: name, Line: line, TextLine: line + 1, Text: text, Entries: []BruEntry{}}
}

func (b BruBlock) Value(key string) string {
	for _, entry := range b.Entries {
		if entry.Key == key && entry.Enabled {
			return entry.Value
		}
	}
	return ""
}

func (b BruBlock) Lookup(key string) (BruEntry, bool) {
	for _, entry := range b.Entries {
		if entry.Key == key {
			return entry, true
		}
	}
	return BruEntry{}, false
}
