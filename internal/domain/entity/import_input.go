package entity

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
)

var xenonImportKinds = []enums.ImportKind{importkind.XenonWorkspaceZip, importkind.XenonWorkspaceFolder, importkind.XenonBundle, importkind.XenonZip, importkind.XenonRequest, importkind.XenonEnvironment}

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

func (i ImportInput) IsPostmanV3() bool {
	return i.Version == PostmanV3Version
}

func (i ImportInput) IsPostmanCollection() bool {
	return i.Kind == importkind.PostmanCollection
}

func (i ImportInput) IsPostmanWorkspace() bool {
	return i.Kind == importkind.PostmanWorkspace
}

func (i ImportInput) IsPostmanRequest() bool {
	return i.Kind == importkind.PostmanRequest
}

func (i ImportInput) IsPostmanEnvironment() bool {
	return i.Kind == importkind.PostmanEnvironment
}

func (i ImportInput) IsPostmanGlobals() bool {
	return i.Kind == importkind.PostmanGlobals
}

func (i ImportInput) IsPostmanDump() bool {
	return i.Kind == importkind.PostmanDump
}

func (i ImportInput) IsBruno() bool {
	return i.Kind == importkind.BrunoBru || i.Kind == importkind.BrunoYaml
}

func (i ImportInput) IsXenon() bool {
	return slices.Contains(xenonImportKinds, i.Kind)
}

func (i ImportInput) IsFolder() bool {
	return i.Probe.IsFolder
}

func (i ImportInput) Entries() []string {
	return i.Probe.Entries
}

func (i ImportInput) Name() string {
	return i.Probe.Name
}

func (i ImportInput) PastedText() string {
	return i.Probe.Text
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
