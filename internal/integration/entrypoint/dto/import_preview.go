package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportPreview struct {
	SourceId string              `json:"source_id"`
	Kind     string              `json:"kind"`
	Version  string              `json:"version"`
	Items    []ImportPreviewItem `json:"items"`
	Warnings []ImportWarning     `json:"warnings"`
	PmUsages []PmUsage           `json:"pm_usages"`
}

func NewImportPreview(source entity.ImportSource) ImportPreview {
	preview := ImportPreview{
		SourceId: source.ID,
		Kind:     source.Kind.String(),
		Version:  source.Version,
		Items:    make([]ImportPreviewItem, 0, len(source.Items)),
		Warnings: NewImportWarnings(source.Warnings()),
		PmUsages: NewPmUsages(source.PmUsages()),
	}
	for _, item := range source.Items {
		preview.Items = append(preview.Items, NewImportPreviewItem(item))
	}
	return preview
}

func (i ImportPreview) LogValue() any {
	return map[string]any{"source_id": i.SourceId, "kind": i.Kind, "items": len(i.Items), "warnings": len(i.Warnings)}
}
