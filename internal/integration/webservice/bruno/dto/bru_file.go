package dto

type BruFile struct {
	Path       string
	Blocks     []BruBlock
	Attributes []BruEntry
}

func NewBruFile(path string) BruFile {
	return BruFile{Path: path, Blocks: []BruBlock{}, Attributes: []BruEntry{}}
}

func (b BruFile) Block(name string) (BruBlock, bool) {
	for _, block := range b.Blocks {
		if block.Name == name {
			return block, true
		}
	}
	return NewBruBlock("", 0), false
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
