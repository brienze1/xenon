package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runtype"
	"github.com/brienze1/xenon/pkg/errs"
)

const NoEnvironmentID = "none"

type RunConfig struct {
	Type          enums.RunType
	EnvironmentID string
	Rows          []RunRow
	Functional    FunctionalSettings
	Performance   PerformanceSettings
}

func NewRunConfig() RunConfig {
	return RunConfig{Type: runtype.Functional, Rows: []RunRow{}, Functional: NewFunctionalSettings(), Performance: NewPerformanceSettings()}
}

func (r RunConfig) IsPerformance() bool {
	return r.Type == runtype.Performance
}

// DataFile is the iteration data file of the run type, empty when none.
func (r RunConfig) DataFile() string {
	if r.IsPerformance() {
		return r.Performance.DataFile
	}
	return r.Functional.DataFile
}

func (r RunConfig) DataFiles() []string {
	files := make([]string, 0, 2)
	for _, file := range []string{r.Functional.DataFile, r.Performance.DataFile} {
		if file != "" {
			files = append(files, file)
		}
	}
	return files
}

func (r RunConfig) WithoutDataFile(path string) RunConfig {
	if r.Functional.DataFile == path {
		r.Functional.DataFile = ""
	}
	if r.Performance.DataFile == path {
		r.Performance.DataFile = ""
	}
	return r
}

func (r RunConfig) WithRows(rows []RunPlanRow) RunConfig {
	r.Rows = make([]RunRow, 0, len(rows))
	for _, row := range rows {
		r.Rows = append(r.Rows, row.RunRow)
	}
	return r
}

// Completed fills the settings a saved configuration left out with Postman's defaults.
func (r RunConfig) Completed() RunConfig {
	defaults := NewRunConfig()
	if r.Type == "" {
		r.Type = defaults.Type
	}
	if r.Functional.Iterations == 0 {
		r.Functional.Iterations = defaults.Functional.Iterations
	}
	if r.Performance.VirtualUsers == 0 && r.Performance.DurationSeconds == 0 {
		r.Performance = NewPerformanceSettings()
	}
	if r.Rows == nil {
		r.Rows = []RunRow{}
	}
	return r
}

func (r RunConfig) SelectedRows() []RunRow {
	selected := make([]RunRow, 0, len(r.Rows))
	for _, row := range r.Rows {
		if row.Selected {
			selected = append(selected, row)
		}
	}
	return selected
}

// Validate checks the selection first, then the settings of the run type.
func (r RunConfig) Validate() error {
	if len(r.SelectedRows()) == 0 {
		return errs.BadRequestError("select at least one request to run", "RUN-03400")
	}
	if r.IsPerformance() {
		return r.Performance.Validate()
	}
	return r.Functional.Validate()
}

func invalidRunSetting(field, reason string) error {
	return errs.BadRequestError("the run setting "+field+" "+reason, "RUN-04400", errs.ErrorDetails{Attribute: field, Messages: []string{reason}})
}
