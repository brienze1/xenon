package dto

type BruEntry struct {
	Key     string
	Value   string
	Values  []string
	Line    int
	Enabled bool
}

func NewBruEntry(key, value string, line int, enabled bool) BruEntry {
	return BruEntry{Key: key, Value: value, Line: line, Enabled: enabled}
}
