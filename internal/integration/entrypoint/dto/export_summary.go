package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ExportSummary struct {
	Secrets  string          `json:"secrets"`
	Path     string          `json:"path"`
	Dropped  []ExportDrop    `json:"dropped"`
	Warnings []ExportWarning `json:"warnings"`
	Counts   ExportCounts    `json:"counts"`
}

func NewExportSummary(summary entity.ExportSummary) ExportSummary {
	dropped := make([]ExportDrop, 0, len(summary.Dropped))
	for _, drop := range summary.Dropped {
		dropped = append(dropped, NewExportDrop(drop))
	}
	warnings := make([]ExportWarning, 0, len(summary.Warnings))
	for _, warning := range summary.Warnings {
		warnings = append(warnings, NewExportWarning(warning))
	}
	return ExportSummary{Secrets: summary.Secrets.String(), Path: summary.Path, Dropped: dropped, Warnings: warnings, Counts: NewExportCounts(summary.Counts)}
}
