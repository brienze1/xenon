package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportCounts struct {
	Collections  int `json:"collections"`
	Folders      int `json:"folders"`
	Requests     int `json:"requests"`
	Examples     int `json:"examples"`
	Variables    int `json:"variables"`
	Scripts      int `json:"scripts"`
	Environments int `json:"environments"`
}

func NewImportCounts(counts entity.ImportCounts) ImportCounts {
	return ImportCounts{
		Collections:  counts.Collections,
		Folders:      counts.Folders,
		Requests:     counts.Requests,
		Examples:     counts.Examples,
		Variables:    counts.Variables,
		Scripts:      counts.Scripts,
		Environments: counts.Environments,
	}
}
