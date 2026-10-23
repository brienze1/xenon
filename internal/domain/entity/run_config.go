package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
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

func (r RunConfig) HasDataFile() bool {
	return r.DataFile() != ""
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
	if r.Type == runtype.Unset {
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

// WithProfileDefaults fills the load profile parameters a performance run left at zero.
func (r RunConfig) WithProfileDefaults() RunConfig {
	if r.IsPerformance() {
		r.Performance = r.Performance.WithProfileDefaults()
	}
	return r
}

func (r RunConfig) WithDataIterations(rows int) RunConfig {
	if !r.IsPerformance() {
		r.Functional = r.Functional.WithDataIterations(rows)
	}
	return r
}

// PlannedRequests is rows × iterations for a functional run; a performance run plans none ahead.
func (r RunConfig) PlannedRequests(rows int) int {
	if r.IsPerformance() {
		return 0
	}
	return rows * r.Functional.Iterations
}

// Override lays the overrides over the configuration; the environment is resolved by the run plan.
func (r RunConfig) Override(overrides RunOverrides) RunConfig {
	if overrides.OverridesRows() {
		r.Rows = overrides.Rows
	}
	r.Type = overridden(overrides.Type, r.Type)
	r.Functional = r.Functional.Override(overrides.Functional)
	r.Performance = r.Performance.Override(overrides.Performance)
	return r
}

// AIChanges are the stored data an AI-started run changes: cookies and variables, only for a functional run.
func (r RunConfig) AIChanges() []enums.AIChangeKind {
	changes := make([]enums.AIChangeKind, 0, 2)
	if r.SavesCookies() {
		changes = append(changes, aichangekind.Cookies)
	}
	if r.KeepsVariableValues() {
		changes = append(changes, aichangekind.Variables)
	}
	return changes
}

// SavesCookies is true when a functional run merges its cookies back; performance runs never do.
func (r RunConfig) SavesCookies() bool {
	return !r.IsPerformance() && r.Functional.SaveCookies
}

func (r RunConfig) KeepsVariableValues() bool {
	return !r.IsPerformance() && r.Functional.KeepVariableValues
}

func (r RunConfig) SeedsStoredCookies() bool {
	return !r.Functional.SkipStoredCookies
}

func (r RunConfig) LogsToConsole() bool {
	return !r.Functional.DisableLogs
}

// EnvironmentIn is the run's environment, else the active one of local, else NoEnvironmentID.
func (r RunConfig) EnvironmentIn(local LocalVariables) string {
	switch {
	case r.EnvironmentID != "":
		return r.EnvironmentID
	case local.HasActiveEnvironment():
		return local.ActiveEnvironmentID
	default:
		return NoEnvironmentID
	}
}

func UsesActiveEnvironment(environmentId string) bool {
	return environmentId == ""
}

func UsesNoEnvironment(environmentId string) bool {
	return environmentId == NoEnvironmentID
}

func (r RunConfig) SelectedRows() []RunRow {
	selected := make([]RunRow, 0, len(r.Rows))
	for _, row := range r.Rows {
		if row.IsSelected() {
			selected = append(selected, row)
		}
	}
	return selected
}

// PlannedRows are the selected rows as the sidebar lists them, numbered in run order.
func (r RunConfig) PlannedRows(sidebar []RunPlanRow) []RunPlanRow {
	rows := make([]RunPlanRow, 0, len(r.Rows))
	for _, row := range r.SelectedRows() {
		planned, _ := RunPlanRowOf(sidebar, row.RequestID)
		planned = planned.For(row)
		planned.Number = len(rows) + 1
		rows = append(rows, planned)
	}
	return rows
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

func invalidRunSetting(field, reason, code string) error {
	return errs.BadRequestError("the run setting "+field+" "+reason, code, errs.ErrorDetails{Attribute: field, Messages: []string{reason}})
}
