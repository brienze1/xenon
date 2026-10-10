package dto

type BruEntry struct {
	Key     string
	Value   string
	Values  []string
	Line    int
	Enabled bool
}

type BruBlock struct {
	Name     string
	Text     string
	Entries  []BruEntry
	Line     int
	TextLine int
}

type BruFile struct {
	Path       string
	Blocks     []BruBlock
	Attributes []BruEntry
}

func (b BruFile) Block(name string) (BruBlock, bool) {
	for _, block := range b.Blocks {
		if block.Name == name {
			return block, true
		}
	}
	return BruBlock{}, false
}

func (b BruFile) Entries(name string) []BruEntry {
	block, _ := b.Block(name)
	return block.Entries
}

func (b BruFile) Value(blockName, key string) string {
	block, _ := b.Block(blockName)
	return block.Value(key)
}

func (b BruFile) Attribute(key string) string {
	for _, entry := range b.Attributes {
		if entry.Key == key {
			return entry.Value
		}
	}
	return ""
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
