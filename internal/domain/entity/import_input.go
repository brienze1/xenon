package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
)

type ImportInput struct {
	Kind      enums.ImportKind
	Version   string
	Root      string
	Documents []ImportDocument
	Probe     ImportProbe
}

func (i ImportInput) Detected(kind enums.ImportKind, version string) ImportInput {
	i.Kind = kind
	i.Version = version
	return i
}

func (i ImportInput) IsPostmanCollection() bool {
	return i.Kind == importkind.PostmanCollection
}

func (i ImportInput) Document() ImportDocument {
	if len(i.Documents) == 0 {
		return ImportDocument{Path: i.Probe.Name}
	}
	return i.Documents[0]
}

func (i ImportInput) LogValue() any {
	return map[string]any{"root": i.Root, "name": i.Probe.Name, "documents": len(i.Documents)}
}
