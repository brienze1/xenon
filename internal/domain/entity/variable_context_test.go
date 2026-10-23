package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

func TestVariableContextWithIterationDataWinsOverEnvironmentAndLosesToLocal(t *testing.T) {
	environment := NewVariableLayer(variablescope.Environment, []VariableRow{NewVariableRow("user", "env-user"), NewVariableRow("region", "eu")}, nil, nil, false)
	context := NewVariableContext(environment).WithIterationData(map[string]any{"user": "data-user", "count": float64(3), "local": "data-local"}).WithLocal(map[string]any{"local": "script"})
	expectations := map[string]string{"user": "data-user", "count": "3", "local": "script", "region": "eu"}
	for name, want := range expectations {
		if got, found := context.Lookup(name); !found || got.Value != want {
			t.Errorf("Lookup(%q) = %+v, %v; want %q", name, got, found, want)
		}
	}
	if got, _ := context.Lookup("user"); got.Scope != variablescope.Data {
		t.Errorf("a data column resolves from the data scope, got %s", got.Scope)
	}
}

func TestVariableContextKeepsIterationDataUnderScriptScopes(t *testing.T) {
	environment := NewVariableLayer(variablescope.Environment, []VariableRow{NewVariableRow("user", "env-user")}, nil, nil, false)
	scopes := ScriptScopes{Environment: NewScriptScope("dev", []ScriptVariable{NewScriptVariable("user", "script-env")}), Local: NewScriptScope("", []ScriptVariable{NewScriptVariable("token", "t")})}
	context := NewVariableContext(environment).WithIterationData(map[string]any{"user": "data-user"}).WithScriptScopes(scopes)
	if got, _ := context.Lookup("user"); got.Value != "data-user" {
		t.Errorf("the data row stays over the environment after script scopes, got %+v", got)
	}
	if got, _ := context.Lookup("token"); got.Value != "t" || got.Scope != variablescope.Local {
		t.Errorf("local stays first, got %+v", got)
	}
	if len(context.Layers) != 3 || context.Layers[1].Scope != variablescope.Data {
		t.Errorf("layers = %+v", context.Layers)
	}
}

func TestVariableContextReplacesThePreviousIterationData(t *testing.T) {
	context := NewVariableContext().WithIterationData(map[string]any{"user": "first"}).WithIterationData(map[string]any{"user": "second"})
	if len(context.Layers) != 1 {
		t.Fatalf("one data layer expected, got %+v", context.Layers)
	}
	if got, _ := context.Lookup("user"); got.Value != "second" {
		t.Errorf("Lookup(user) = %+v", got)
	}
}
