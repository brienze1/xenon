package entity

import "testing"

func TestEnvironmentListIsActive(t *testing.T) {
	list := EnvironmentList{ActiveID: "e1"}
	if !list.IsActive("e1") || list.IsActive("e2") {
		t.Errorf("IsActive mismatch for %+v", list)
	}
}

func TestEnvironmentListWithActive(t *testing.T) {
	list := EnvironmentList{ActiveID: "e1", Environments: []Environment{{ID: "e1"}, {ID: "e2"}}}
	tests := map[string]string{"e2": "e2", NoEnvironmentID: "", "deleted": "", "": ""}
	for id, want := range tests {
		if got := list.WithActive(id).ActiveID; got != want {
			t.Errorf("WithActive(%q).ActiveID = %q, want %q", id, got, want)
		}
	}
	if list.ActiveID != "e1" {
		t.Errorf("list changed: %q", list.ActiveID)
	}
}
