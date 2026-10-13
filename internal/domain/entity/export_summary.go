package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ExportSummary struct {
	Secrets  enums.ExportSecrets
	Path     string
	Dropped  []ExportDrop
	Warnings []ExportWarning
	Counts   ExportCounts
}

func NewExportSummary(path string, secrets enums.ExportSecrets) ExportSummary {
	return ExportSummary{Path: path, Secrets: secrets, Dropped: []ExportDrop{}, Warnings: []ExportWarning{}}
}

func (e ExportSummary) With(counts ExportCounts, dropped []ExportDrop, warnings []ExportWarning) ExportSummary {
	e.Counts = e.Counts.Add(counts)
	e.Dropped = append(e.Dropped, dropped...)
	e.Warnings = append(e.Warnings, warnings...)
	return e
}
