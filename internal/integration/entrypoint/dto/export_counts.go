package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ExportCounts struct {
	Collections  int `json:"collections"`
	Folders      int `json:"folders"`
	Requests     int `json:"requests"`
	Examples     int `json:"examples"`
	Environments int `json:"environments"`
	Variables    int `json:"variables"`
}

func NewExportCounts(counts entity.ExportCounts) ExportCounts {
	return ExportCounts{
		Collections:  counts.Collections,
		Folders:      counts.Folders,
		Requests:     counts.Requests,
		Examples:     counts.Examples,
		Environments: counts.Environments,
		Variables:    counts.Variables,
	}
}
