package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/loadprofilekind"
	"github.com/brienze1/xenon/internal/domain/enums/runtype"
	"github.com/brienze1/xenon/pkg/errs"
)

func selectedRunConfig() RunConfig {
	config := NewRunConfig()
	config.Rows = []RunRow{{RequestID: "a", Selected: true}, {RequestID: "b"}}
	return config
}

func expectRunError(t *testing.T, err error, code, field string) {
	t.Helper()
	failure, ok := err.(errs.BaseError)
	if !ok || failure.Code() != code {
		t.Fatalf("want %s, got %v", code, err)
	}
	if field == "" {
		return
	}
	if failure.ErrorDetails()["attribute"] != field {
		t.Errorf("want the details to name %s, got %v", field, failure.ErrorDetails())
	}
}

func TestNewRunConfigHasPostmansDefaults(t *testing.T) {
	config := NewRunConfig()
	if config.Type != runtype.Functional || config.Functional.Iterations != 1 || config.Functional.DelayMs != 0 || !config.Functional.KeepVariableValues {
		t.Errorf("functional defaults = %+v", config.Functional)
	}
	if config.Functional.PersistResponses || config.Functional.DisableLogs || config.Functional.StopOnError || config.Functional.SkipStoredCookies || config.Functional.SaveCookies {
		t.Errorf("functional switches default off, got %+v", config.Functional)
	}
	if config.Performance.VirtualUsers != 10 || config.Performance.DurationSeconds != 600 || config.Performance.Profile.Kind != loadprofilekind.Fixed {
		t.Errorf("performance defaults = %+v", config.Performance)
	}
	if config.EnvironmentID != "" || len(config.Rows) != 0 {
		t.Errorf("config = %+v", config)
	}
}

func TestRunConfigSelectedRowsKeepsTheOrder(t *testing.T) {
	config := NewRunConfig()
	config.Rows = []RunRow{{RequestID: "c", Selected: true}, {RequestID: "a"}, {RequestID: "b", Selected: true}}
	selected := config.SelectedRows()
	if len(selected) != 2 || selected[0].RequestID != "c" || selected[1].RequestID != "b" {
		t.Errorf("selected = %+v", selected)
	}
}

func TestRunConfigValidateRejectsAnEmptySelection(t *testing.T) {
	config := NewRunConfig()
	config.Rows = []RunRow{{RequestID: "a"}}
	expectRunError(t, config.Validate(), "RUN-03400", "")
	if err := selectedRunConfig().Validate(); err != nil {
		t.Errorf("a valid config: %v", err)
	}
}

func TestRunConfigValidateRejectsFunctionalSettingsOutOfRange(t *testing.T) {
	cases := map[string]func(*FunctionalSettings){
		"iterations": func(s *FunctionalSettings) { s.Iterations = 0 },
		"delay_ms":   func(s *FunctionalSettings) { s.DelayMs = -1 },
	}
	for field, breakSettings := range cases {
		config := selectedRunConfig()
		breakSettings(&config.Functional)
		expectRunError(t, config.Validate(), "RUN-04400", field)
	}
	config := selectedRunConfig()
	config.Functional.DelayMs = 3600001
	expectRunError(t, config.Validate(), "RUN-04400", "delay_ms")
	config.Functional.DelayMs = 3600000
	if err := config.Validate(); err != nil {
		t.Errorf("the maximum delay is valid: %v", err)
	}
}

func TestRunConfigValidateRejectsPerformanceSettingsOutOfRange(t *testing.T) {
	cases := []struct {
		breakSettings func(*PerformanceSettings)
		field         string
	}{
		{func(s *PerformanceSettings) { s.VirtualUsers = 0 }, "virtual_users"},
		{func(s *PerformanceSettings) { s.VirtualUsers = 101 }, "virtual_users"},
		{func(s *PerformanceSettings) { s.DurationSeconds = 0 }, "duration_seconds"},
		{func(s *PerformanceSettings) { s.DurationSeconds = 3601 }, "duration_seconds"},
		{func(s *PerformanceSettings) {
			s.VirtualUsers, s.Profile = 4, LoadProfile{Kind: loadprofilekind.RampUp, InitialUsers: 5, RampSeconds: 1}
		}, "profile.initial_users"},
		{func(s *PerformanceSettings) {
			s.DurationSeconds, s.Profile = 3, LoadProfile{Kind: loadprofilekind.RampUp, InitialUsers: 1, RampSeconds: 4}
		}, "profile.ramp_seconds"},
		{func(s *PerformanceSettings) {
			s.DurationSeconds, s.Profile = 3, LoadProfile{Kind: loadprofilekind.Spike, BaseUsers: 1, SpikeSeconds: 4}
		}, "profile.spike_seconds"},
		{func(s *PerformanceSettings) {
			s.VirtualUsers, s.Profile = 4, LoadProfile{Kind: loadprofilekind.Spike, BaseUsers: 5, SpikeSeconds: 1}
		}, "profile.base_users"},
		{func(s *PerformanceSettings) {
			s.DurationSeconds, s.Profile = 3, LoadProfile{Kind: loadprofilekind.Peak, BaseUsers: 1, PeakSeconds: 3}
		}, "profile.peak_seconds"},
	}
	for _, testCase := range cases {
		config := selectedRunConfig()
		config.Type = runtype.Performance
		testCase.breakSettings(&config.Performance)
		expectRunError(t, config.Validate(), "RUN-04400", testCase.field)
	}
	config := selectedRunConfig()
	config.Type = runtype.Performance
	config.Functional.Iterations = 0
	if err := config.Validate(); err != nil {
		t.Errorf("a performance run ignores the functional settings: %v", err)
	}
}
