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
