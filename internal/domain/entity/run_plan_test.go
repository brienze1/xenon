package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/loadprofilekind"
	"github.com/brienze1/xenon/internal/domain/enums/runtype"
)

func rememberedRunPlan() RunPlan {
	return RunPlan{
		Config: RunConfig{
			Type:          runtype.Functional,
			EnvironmentID: "dev-id",
			Rows:          []RunRow{{RequestID: "a", Selected: true}, {RequestID: "b", Selected: false}},
			Functional:    FunctionalSettings{Iterations: 3, DelayMs: 10, KeepVariableValues: true},
			Performance:   PerformanceSettings{VirtualUsers: 10, DurationSeconds: 600, Profile: LoadProfile{Kind: loadprofilekind.Fixed}},
		},
		Environments: []Environment{{ID: "dev-id", Name: "dev"}, {ID: "stage-id", Name: "stage"}},
	}
}

func pointerTo[T any](value T) *T {
	return &value
}

func TestRunPlanKeepsTheRememberedConfigWithoutOverrides(t *testing.T) {
	config, known := rememberedRunPlan().Configured(RunOverrides{})
	if !known || config.EnvironmentID != "dev-id" || len(config.Rows) != 2 || config.Functional.Iterations != 3 {
		t.Errorf("config = %+v", config)
	}
}

func TestRunPlanLaysTheOverridesOverTheRememberedConfig(t *testing.T) {
	overrides := RunOverrides{
		Rows:        []RunRow{{RequestID: "b", AliasID: "x", Selected: true}},
		Type:        pointerTo(runtype.Performance),
		Functional:  FunctionalOverrides{DelayMs: pointerTo(0)},
		Performance: PerformanceOverrides{VirtualUsers: pointerTo(2), Profile: LoadProfileOverrides{Kind: pointerTo(loadprofilekind.Spike)}},
		Environment: pointerTo("stage"),
	}
	config, known := rememberedRunPlan().Configured(overrides)
	if !known || config.EnvironmentID != "stage-id" || config.Type != runtype.Performance {
		t.Errorf("config = %+v", config)
	}
	if len(config.Rows) != 1 || config.Rows[0] != (RunRow{RequestID: "b", AliasID: "x", Selected: true}) {
		t.Errorf("rows = %+v", config.Rows)
	}
	if config.Functional.DelayMs != 0 || config.Functional.Iterations != 3 {
		t.Errorf("functional = %+v", config.Functional)
	}
	if config.Performance.VirtualUsers != 2 || config.Performance.DurationSeconds != 600 || config.Performance.Profile.Kind != loadprofilekind.Spike {
		t.Errorf("performance = %+v", config.Performance)
	}
}

func TestRunPlanIteratesOverADataFileWithoutAnIterationCount(t *testing.T) {
	config, _ := rememberedRunPlan().Configured(RunOverrides{Functional: FunctionalOverrides{DataFile: pointerTo("/d.csv")}})
	if config.Functional.Iterations != 0 || config.Functional.DataFile != "/d.csv" {
		t.Errorf("functional = %+v", config.Functional)
	}
	config, _ = rememberedRunPlan().Configured(RunOverrides{Functional: FunctionalOverrides{DataFile: pointerTo("/d.csv"), Iterations: pointerTo(2)}})
	if config.Functional.Iterations != 2 {
		t.Errorf("functional = %+v", config.Functional)
	}
}

func TestRunPlanEnvironment(t *testing.T) {
	config, known := rememberedRunPlan().Configured(RunOverrides{Environment: pointerTo(NoEnvironmentID)})
	if !known || config.EnvironmentID != NoEnvironmentID {
		t.Errorf("config = %+v", config)
	}
	if _, known := rememberedRunPlan().Configured(RunOverrides{Environment: pointerTo("nope")}); known {
		t.Error("an unknown environment must be reported")
	}
}

func TestRunConfigAIChanges(t *testing.T) {
	plan := rememberedRunPlan()
	saveCookies := FunctionalOverrides{SaveCookies: pointerTo(true)}
	config, _ := plan.Configured(RunOverrides{Functional: saveCookies})
	if changes := config.AIChanges(); len(changes) != 2 {
		t.Errorf("changes = %v", changes)
	}
	config, _ = plan.Configured(RunOverrides{Type: pointerTo(runtype.Performance), Functional: saveCookies})
	if changes := config.AIChanges(); len(changes) != 0 {
		t.Errorf("a performance run changes nothing, got %v", changes)
	}
}
