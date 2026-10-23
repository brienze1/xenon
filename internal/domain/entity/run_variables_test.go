package entity

import "testing"

func TestRunVariablesOverlayShowsAChangeThroughOver(t *testing.T) {
	stored := ScriptScopes{Environment: NewScriptScope("dev", []ScriptVariable{NewScriptVariable("token", "old")}), Globals: NewScriptScope("", nil)}
	after := stored
	after.Environment = after.Environment.WithValue("token", "new").WithValue("added", "yes")
	overlay := NewRunVariables()
	overlay.Apply(stored, after, nil, nil)
	seen := overlay.Over(stored)
	if value, _ := seen.Environment.Lookup("token"); value.Text() != "new" {
		t.Errorf("token = %+v", value)
	}
	if value, found := seen.Environment.Lookup("added"); !found || value.Text() != "yes" {
		t.Errorf("added = %+v, %v", value, found)
	}
	if value, _ := stored.Environment.Lookup("token"); value.Text() != "old" {
		t.Errorf("the stored scopes are not touched, got %+v", value)
	}
}

func TestRunVariablesOverlayUnsetRemovesTheValue(t *testing.T) {
	stored := ScriptScopes{Collection: NewScriptScope("Shop", []ScriptVariable{NewScriptVariable("cursor", "1"), NewScriptVariable("kept", "k")})}
	after := stored
	after.Collection = after.Collection.WithoutValue("cursor")
	overlay := NewRunVariables()
	overlay.Apply(stored, after, nil, nil)
	seen := overlay.Over(stored)
	if _, found := seen.Collection.Lookup("cursor"); found {
		t.Error("an unset value is gone in later rows")
	}
	if _, found := seen.Collection.Lookup("kept"); !found {
		t.Error("other values stay")
	}
	again := seen
	again.Collection = again.Collection.WithValue("cursor", "2")
	overlay.Apply(seen, again, nil, nil)
	if value, found := overlay.Over(stored).Collection.Lookup("cursor"); !found || value.Text() != "2" {
		t.Errorf("a value set after an unset is back, got %+v, %v", value, found)
	}
}

func TestRunVariablesOverlayTracksVaultChanges(t *testing.T) {
	overlay := NewRunVariables()
	overlay.Apply(ScriptScopes{}, ScriptScopes{}, map[string]string{"apiKey": "a", "old": "o"}, map[string]string{"apiKey": "b", "fresh": "f"})
	vault := overlay.OverVault(map[string]string{"apiKey": "a", "old": "o", "other": "x"})
	if vault["apiKey"] != "b" || vault["fresh"] != "f" || vault["other"] != "x" {
		t.Errorf("vault = %v", vault)
	}
	if _, found := vault["old"]; found {
		t.Errorf("a removed vault secret stays removed, got %v", vault)
	}
	if empty := overlay.OverVault(nil); empty["apiKey"] != "b" {
		t.Errorf("an empty vault still sees the run's secrets, got %v", empty)
	}
}

func TestRunContextDefaultsToASingleSend(t *testing.T) {
	single := NoRunContext()
	if single.IsRun() || !single.KeepsVariables() || single.JarID("ws") != "ws" || !single.UsesWorkspaceJar("ws") || single.Collects() {
		t.Errorf("single = %+v", single)
	}
	if info := single.ScriptInfo(NewScriptInfo("r", "R", "R", nil)); info.Iteration != 0 || info.IterationCount != 1 {
		t.Errorf("info = %+v", info)
	}
	if RunCookieJarID("ws", "run1") != "ws#run:run1" || RunResponseKey("run1", 3) != "run:run1:3" {
		t.Errorf("ids = %s, %s", RunCookieJarID("ws", "run1"), RunResponseKey("run1", 3))
	}
	run := RunContext{RunID: "run1", CookieJarID: RunCookieJarID("ws", "run1"), Variables: NewRunVariables(), Iteration: 2, IterationCount: 3}
	if !run.IsRun() || run.KeepsVariables() || run.UsesWorkspaceJar("ws") {
		t.Errorf("run = %+v", run)
	}
	if info := run.ScriptInfo(NewScriptInfo("r", "R", "R", nil)); info.Iteration != 2 || info.IterationCount != 3 {
		t.Errorf("info = %+v", info)
	}
}
