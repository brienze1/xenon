package entity

import "testing"

func TestNewAISessionContextSeedsTheWorkspaceAndItsEnvironment(t *testing.T) {
	tests := []struct {
		name            string
		localId         string
		storedActive    string
		wantEnvironment string
		wantEntered     bool
	}{
		{name: "workspace with an active environment", localId: "w1", storedActive: "dev", wantEnvironment: "dev", wantEntered: true},
		{name: "workspace without an active environment", localId: "w1", storedActive: "", wantEnvironment: NoEnvironmentID, wantEntered: true},
		{name: "no workspace", localId: "", storedActive: "dev", wantEnvironment: NoEnvironmentID, wantEntered: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context := NewAISessionContext("s1", test.localId, test.storedActive)
			if context.SessionID != "s1" || context.LocalID != test.localId {
				t.Errorf("context = %+v", context)
			}
			if got := context.EnvironmentID(); got != test.wantEnvironment {
				t.Errorf("EnvironmentID() = %q, want %q", got, test.wantEnvironment)
			}
			if got := context.HasEntered(test.localId); got != test.wantEntered {
				t.Errorf("HasEntered(%q) = %v, want %v", test.localId, got, test.wantEntered)
			}
		})
	}
}

func TestAISessionContextEnterKeepsTheChoiceMadeOnAnEarlierVisit(t *testing.T) {
	context := NewAISessionContext("s1", "w1", "dev").ChooseEnvironment("prod").Enter("w2", "qa")
	if context.LocalID != "w2" || context.EnvironmentID() != "qa" {
		t.Fatalf("after entering w2: %+v", context)
	}
	back := context.Enter("w1", "dev")
	if back.LocalID != "w1" || back.EnvironmentID() != "prod" {
		t.Errorf("re-entering w1 = %+v, want prod kept", back)
	}
}

func TestAISessionContextChooseEnvironment(t *testing.T) {
	tests := []struct {
		name   string
		chosen string
		want   string
	}{
		{name: "an environment", chosen: "prod", want: "prod"},
		{name: "empty means none", chosen: "", want: NoEnvironmentID},
		{name: "none", chosen: NoEnvironmentID, want: NoEnvironmentID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NewAISessionContext("s1", "w1", "dev").ChooseEnvironment(test.chosen).EnvironmentID(); got != test.want {
				t.Errorf("EnvironmentID() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestAISessionContextNeverSharesItsChoicesWithACopy(t *testing.T) {
	original := NewAISessionContext("s1", "w1", "dev")
	original.ChooseEnvironment("prod")
	original.Enter("w2", "qa")
	if original.EnvironmentID() != "dev" || original.HasEntered("w2") {
		t.Errorf("original changed: %+v", original)
	}
}

func TestAISessionContextLeaveDropsTheWorkspaceAndKeepsTheChoices(t *testing.T) {
	left := NewAISessionContext("s1", "w1", "dev").ChooseEnvironment("prod").Leave()
	if left.HasWorkspace() || left.EnvironmentID() != NoEnvironmentID {
		t.Errorf("left = %+v", left)
	}
	if left.ChooseEnvironment("dev").HasEntered("") {
		t.Errorf("choosing without a workspace recorded a choice")
	}
	if got := left.Enter("w1", "dev").EnvironmentID(); got != "prod" {
		t.Errorf("re-entering after leaving = %q, want prod", got)
	}
}

func TestAISessionContextActiveEnvironmentID(t *testing.T) {
	if got := NewAISessionContext("s1", "w1", "dev").ActiveEnvironmentID(); got != "dev" {
		t.Errorf("ActiveEnvironmentID() = %q, want dev", got)
	}
	none := NewAISessionContext("s1", "w1", "dev").ChooseEnvironment("")
	if none.HasEnvironment() || none.ActiveEnvironmentID() != "" {
		t.Errorf("none = %+v", none)
	}
}

func TestAISessionContextInRegistryLeavesAnUnregisteredWorkspace(t *testing.T) {
	registry := NewAppRegistry()
	registry.Register(RegisteredWorkspace{LocalID: "w1"})
	if kept := NewAISessionContext("s1", "w1", "dev").InRegistry(&registry); kept.LocalID != "w1" {
		t.Errorf("registered workspace left: %+v", kept)
	}
	if left := NewAISessionContext("s1", "w2", "dev").InRegistry(&registry); left.HasWorkspace() {
		t.Errorf("unregistered workspace kept: %+v", left)
	}
	if none := NewAISessionContext("s1", "", "").InRegistry(&registry); none.HasWorkspace() {
		t.Errorf("no workspace became one: %+v", none)
	}
}
