package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/clashresolution"
)

type MCPImport struct {
	SourceId   string   `json:"source_id" description:"Source id returned by detect; omitted means nothing to discard." used_by:"discard"`
	OnClash    string   `json:"on_clash" description:"How to settle a name clash; default keep_both on run." used_by:"clashes,run" enum:"keep_both,replace,skip" validated_as:"apply_to_all,items[].resolution"`
	Passphrase string   `json:"passphrase" description:"Passphrase of an export with encrypted secrets." used_by:"clashes,run"`
	Paths      []string `json:"paths" description:"Absolute paths of the files or folders to import." used_by:"detect,clashes,run" required_for:"detect,clashes,run" validated_as:"text"`
}

func (m *MCPImport) DetectImport() DetectImport {
	return NewDetectImport(m.Paths)
}

func (m *MCPImport) Discard() ImportSourceReference {
	return NewImportSourceReference(m.SourceId)
}

func (m *MCPImport) HasSource() bool {
	return m.SourceId != ""
}

// WithDefaultClash keeps both copies when the client did not say how to settle a clash.
func (m MCPImport) WithDefaultClash() MCPImport {
	if m.OnClash == "" {
		m.OnClash = clashresolution.KeepBoth.String()
	}
	return m
}

func (m *MCPImport) ImportPlan(localId string, source entity.ImportSource) ImportPlan {
	items := make([]ImportPlanItem, 0, len(source.Items))
	for _, item := range source.Items {
		items = append(items, NewImportPlanItem(item.Key, localId, m.OnClash, item.IncludedInAIImport()))
	}
	return NewImportPlan(source.ID, m.OnClash, m.Passphrase, items)
}

func (m MCPImport) LogValue() any {
	return map[string]any{"source_id": m.SourceId, "on_clash": m.OnClash, "passphrase_length": len(m.Passphrase), "paths": m.Paths}
}
