package entity

import "testing"

func TestEnvironmentBlocksSave(t *testing.T) {
	stored := Environment{ID: "01E", Revision: "abc"}
	cases := []struct {
		name         string
		baseRevision string
		force        bool
		blocks       bool
	}{
		{name: "current revision", baseRevision: "abc"},
		{name: "outdated revision", baseRevision: "old", blocks: true},
		{name: "outdated but forced", baseRevision: "old", force: true},
		{name: "no revision patches the latest", baseRevision: ""},
	}
	for _, c := range cases {
		if got := stored.BlocksSave(c.baseRevision, c.force); got != c.blocks {
			t.Errorf("%s: BlocksSave() = %v, want %v", c.name, got, c.blocks)
		}
	}
}

func TestEnvironmentDuplicatedDropsTheRevision(t *testing.T) {
	if copied := (Environment{ID: "01E", Revision: "abc"}).Duplicated("01F", "copy"); copied.Revision != "" {
		t.Fatalf("Duplicated() revision = %q", copied.Revision)
	}
}

func TestVariableTableBlocksSaveOnlyGuardsEnvironments(t *testing.T) {
	environment := VariableTable{Scope: NewEnvironmentScopeRef("01E")}.WithRevision("abc")
	if !environment.BlocksSave("old", false) || environment.BlocksSave("abc", false) || environment.BlocksSave("old", true) || environment.BlocksSave("", false) {
		t.Error("an environment table blocks only an outdated, unforced revision")
	}
	collection := VariableTable{Scope: NewCollectionScopeRef("01C")}
	if collection.BlocksSave("old", false) {
		t.Error("a collection table never blocks a save")
	}
}

func TestImportPlanTargetLocalIDs(t *testing.T) {
	plan := ImportPlan{Items: []ImportPlanItem{
		{TargetLocalID: "ws-2", Include: true},
		{TargetLocalID: "ws-1", Include: true},
		{TargetLocalID: "ws-2", Include: true},
		{TargetLocalID: "ws-3"},
		{Include: true},
	}}
	got := plan.TargetLocalIDs()
	if len(got) != 2 || got[0] != "ws-1" || got[1] != "ws-2" {
		t.Fatalf("TargetLocalIDs() = %v", got)
	}
}
