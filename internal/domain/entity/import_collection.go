package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ImportCollection struct {
	Secrets  map[string]string
	Warnings []ImportWarning
	Root     ImportNode
}

func NewImportCollection(root ImportNode, warnings []ImportWarning, secrets map[string]string) ImportCollection {
	if secrets == nil {
		secrets = map[string]string{}
	}
	if warnings == nil {
		warnings = []ImportWarning{}
	}
	return ImportCollection{Root: root, Warnings: warnings, Secrets: secrets}
}

func (i ImportCollection) ImportItem(key string, kind enums.ImportItemKind) ImportItem {
	return NewImportItem(key, kind, i.Root, i.Warnings).WithCollectionSecrets(i.Secrets)
}

func (i ImportCollection) FirstChild() (ImportCollection, bool) {
	if len(i.Root.Children) == 0 {
		return ImportCollection{}, false
	}
	return NewImportCollection(i.Root.Children[0], i.Warnings, i.Secrets), true
}

func (i ImportCollection) WithItemWarnings(warnings []ImportWarning) ImportCollection {
	ids := i.Root.IDsByPath("")
	for _, warning := range warnings {
		i.Warnings = append(i.Warnings, warning.OnItem(ids[warning.Path]))
	}
	return i
}

func (i ImportCollection) LogValue() any {
	return map[string]any{"name": i.Root.Name(), "warnings": len(i.Warnings), "secrets": len(i.Secrets)}
}
