package entity

import "github.com/brienze1/xenon/internal/domain/enums"

// RunOverrides are the settings laid over a remembered run configuration; nil keeps the remembered value.
type RunOverrides struct {
	Functional  FunctionalOverrides
	Performance PerformanceOverrides
	Type        *enums.RunType
	Environment *string
	Rows        []RunRow
}

func (r RunOverrides) OverridesRows() bool {
	return r.Rows != nil
}

func (r RunOverrides) OverridesEnvironment() bool {
	return r.Environment != nil
}

func overridden[T any](value *T, fallback T) T {
	if value == nil {
		return fallback
	}
	return *value
}
