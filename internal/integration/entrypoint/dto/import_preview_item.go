package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportPreviewItem struct {
	Key           string       `json:"key"`
	Kind          string       `json:"kind"`
	Name          string       `json:"name"`
	Counts        ImportCounts `json:"counts"`
	WarningsCount int          `json:"warnings_count"`
}

func NewImportPreviewItem(item entity.ImportItem) ImportPreviewItem {
	return ImportPreviewItem{
		Key:           item.Key,
		Kind:          item.Kind.String(),
		Name:          item.Name,
		Counts:        NewImportCounts(item.Counts()),
		WarningsCount: len(item.Warnings),
	}
}
