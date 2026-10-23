package dto

import (
	"encoding/json"
	"testing"
)

func decodedMCPRun(t *testing.T, raw string) MCPRun {
	t.Helper()
	var arguments MCPRun
	if err := json.Unmarshal([]byte(raw), &arguments); err != nil {
		t.Fatal(err)
	}
	return arguments
}

func rememberedRunPlan() RunPlan {
	return RunPlan{
		Config: RunConfig{
			Type:          "functional",
			EnvironmentId: "dev-id",
			Rows:          []RunRow{{RequestId: "a", Selected: true}, {RequestId: "b", Selected: false}},
			Functional:    RunFunctionalSettings{Iterations: 3, DelayMs: 10, KeepVariableValues: true},
			Performance:   RunPerformanceSettings{VirtualUsers: 10, DurationSeconds: 600, Profile: RunLoadProfile{Kind: "fixed"}},
		},
		Environments: []EnvironmentSummary{{Id: "dev-id", Name: "dev"}, {Id: "stage-id", Name: "stage"}},
	}
}

func TestMCPRunKeepsTheRememberedConfigWithoutArguments(t *testing.T) {
	arguments := decodedMCPRun(t, `{"target":"Shop"}`)
	config, known := arguments.RunConfig(rememberedRunPlan())
	if !known || config.EnvironmentId != "dev-id" || len(config.Rows) != 2 || config.Functional.Iterations != 3 {
		t.Errorf("config = %+v", config)
	}
}

func TestMCPRunLaysTheArgumentsOverTheRememberedConfig(t *testing.T) {
	arguments := decodedMCPRun(t, `{"rows":[{"request_id":"b","alias_id":"x"}],"type":"performance","functional":{"delay_ms":0},"performance":{"virtual_users":2,"profile":{"kind":"spike"}},"environment":"stage"}`)
	config, known := arguments.RunConfig(rememberedRunPlan())
	if !known || config.EnvironmentId != "stage-id" || config.Type != "performance" {
		t.Errorf("config = %+v", config)
	}
	if len(config.Rows) != 1 || config.Rows[0] != (RunRow{RequestId: "b", AliasId: "x", Selected: true}) {
		t.Errorf("rows = %+v", config.Rows)
	}
	if config.Functional.DelayMs != 0 || config.Functional.Iterations != 3 {
		t.Errorf("functional = %+v", config.Functional)
	}
	if config.Performance.VirtualUsers != 2 || config.Performance.DurationSeconds != 600 || config.Performance.Profile.Kind != "spike" {
		t.Errorf("performance = %+v", config.Performance)
	}
}

func TestMCPRunIteratesOverADataFileWithoutAnIterationCount(t *testing.T) {
	withData := decodedMCPRun(t, `{"functional":{"data_file":"/d.csv"}}`)
	config, _ := withData.RunConfig(rememberedRunPlan())
	if config.Functional.Iterations != 0 || config.Functional.DataFile != "/d.csv" {
		t.Errorf("functional = %+v", config.Functional)
	}
	withIterations := decodedMCPRun(t, `{"functional":{"data_file":"/d.csv","iterations":2}}`)
	config, _ = withIterations.RunConfig(rememberedRunPlan())
	if config.Functional.Iterations != 2 {
		t.Errorf("functional = %+v", config.Functional)
	}
}

func TestMCPRunEnvironment(t *testing.T) {
	none := decodedMCPRun(t, `{"environment":"none"}`)
	config, known := none.RunConfig(rememberedRunPlan())
	if !known || config.EnvironmentId != "none" {
		t.Errorf("config = %+v", config)
	}
	unknown := decodedMCPRun(t, `{"environment":"nope"}`)
	if _, known := unknown.RunConfig(rememberedRunPlan()); known {
		t.Error("an unknown environment must be reported")
	}
}

func TestMCPRunChanges(t *testing.T) {
	plan := rememberedRunPlan()
	arguments := decodedMCPRun(t, `{"functional":{"save_cookies":true}}`)
	config, _ := arguments.RunConfig(plan)
	if changes := arguments.Changes(config); len(changes) != 2 {
		t.Errorf("changes = %v", changes)
	}
	arguments = decodedMCPRun(t, `{"type":"performance","functional":{"save_cookies":true}}`)
	config, _ = arguments.RunConfig(plan)
	if changes := arguments.Changes(config); len(changes) != 0 {
		t.Errorf("a performance run changes nothing, got %v", changes)
	}
}
