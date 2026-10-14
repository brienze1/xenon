package entity

import "testing"

func TestScopeRefDisplayName(t *testing.T) {
	cases := map[string]string{"globals": "Globals", "env:e1": "env:e1", "collection:c1": "collection:c1", "bogus": "bogus"}
	for raw, expected := range cases {
		if name := ParseScopeRef(raw).DisplayName(); name != expected {
			t.Errorf("DisplayName(%q) = %q, want %q", raw, name, expected)
		}
	}
}

func TestScopeRefTarget(t *testing.T) {
	if target := NewEnvironmentScopeRef("e1").Target("Dev"); target != (AIActivityTarget{Name: "Dev", Path: "env:e1"}) {
		t.Errorf("Target() = %+v", target)
	}
}
