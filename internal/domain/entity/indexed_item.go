package entity

import (
	"path/filepath"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type IndexedItem struct {
	Path string
	Kind enums.ItemKind
}

func NewIndexedItem(path string, kind enums.ItemKind) IndexedItem {
	return IndexedItem{Path: path, Kind: kind}
}

func (i IndexedItem) IsRequest() bool {
	return i.Kind == itemkind.Request
}

func (i IndexedItem) MovedTo(path string) IndexedItem {
	i.Path = path
	return i
}

func (i IndexedItem) IsInside(dir string) bool {
	relative, err := filepath.Rel(dir, i.Path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
