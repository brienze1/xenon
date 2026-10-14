package entity

import "testing"

func TestEnvironmentTargets(t *testing.T) {
	environment := Environment{ID: "e1", Name: "Dev"}
	if target := environment.Target(); target != (AIActivityTarget{ID: "e1", Name: "Dev"}) {
		t.Errorf("Target() = %+v", target)
	}
	if target := NewAIEnvironmentTarget("e2"); target != (AIActivityTarget{ID: "e2"}) {
		t.Errorf("NewAIEnvironmentTarget() = %+v", target)
	}
}
