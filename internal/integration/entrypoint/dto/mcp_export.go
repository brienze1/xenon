package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
)

const mcpXenonExportFormat = "xenon"

type MCPExport struct {
	Target           MCPExportTarget `json:"target"`
	Format           string          `json:"format"`
	Path             string          `json:"path"`
	CurrentAsInitial bool            `json:"current_as_initial"`
	IncludeSecrets   bool            `json:"include_secrets"`
}

func (m *MCPExport) RequestsSecrets() bool {
	return m.IncludeSecrets
}

func (m *MCPExport) ExportItemReference(localId string) ExportItemReference {
	return NewExportItemReference(localId, m.Target.Kind, m.Target.Id)
}

// ExportFormat accepts "xenon" as a short name for the Xenon YAML format.
func (m *MCPExport) ExportFormat() string {
	if m.Format == mcpXenonExportFormat {
		return exportformat.XenonYaml.String()
	}
	return m.Format
}

func (m *MCPExport) Export(localId string, options entity.ExportOptions) Export {
	format := m.ExportFormat()
	shape := options.ShapeFor(exportformat.GetExportFormat(format), m.Path).String()
	return NewExport(localId, m.Target.Kind, m.Target.Id, format, shape, m.Path, m.CurrentAsInitial)
}

func (m *MCPExport) ActivityTarget() entity.AIActivityTarget {
	return entity.AIActivityTarget{ID: m.Target.Id, Path: m.Path}
}
