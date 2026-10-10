package entity

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const ImportSourceLifetime = 30 * time.Minute

type ImportSource struct {
	ExpiresAt      time.Time
	ID             string
	Kind           enums.ImportKind
	Version        string
	TempDirs       []string
	Items          []ImportItem
	SourceWarnings []ImportWarning
}

func NewImportSource(id string, inputs []ImportInput, tempDirs []string, now time.Time) ImportSource {
	source := ImportSource{ID: id, TempDirs: tempDirs, Items: []ImportItem{}, SourceWarnings: []ImportWarning{}, ExpiresAt: now.Add(ImportSourceLifetime)}
	if len(inputs) > 0 {
		source.Kind = inputs[0].Kind
		source.Version = inputs[0].Version
	}
	return source
}

func (i ImportSource) WithItems(items []ImportItem) ImportSource {
	i.Items = items
	return i
}

func (i ImportSource) WithSourceWarnings(warnings []ImportWarning) ImportSource {
	i.SourceWarnings = append([]ImportWarning{}, warnings...)
	return i
}

func (i ImportSource) Item(key string) (ImportItem, bool) {
	for _, item := range i.Items {
		if item.Key == key {
			return item, true
		}
	}
	return ImportItem{}, false
}

func (i ImportSource) Warnings() []ImportWarning {
	warnings := make([]ImportWarning, 0)
	for _, item := range i.Items {
		warnings = append(warnings, item.Warnings...)
	}
	return append(warnings, i.SourceWarnings...)
}

func (i ImportSource) PmUsages() []PmUsage {
	usages := make([]PmUsage, 0)
	for _, item := range i.Items {
		usages = append(usages, item.PmUsages...)
	}
	return MergePmUsages(usages)
}

func (i ImportSource) IsExpired(now time.Time) bool {
	return !now.Before(i.ExpiresAt)
}

func (i ImportSource) LogValue() any {
	return map[string]any{"id": i.ID, "kind": i.Kind, "version": i.Version, "items": len(i.Items)}
}
