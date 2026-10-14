package entity

import "testing"

func TestEnvironmentListIsActive(t *testing.T) {
	list := EnvironmentList{ActiveID: "e1"}
	if !list.IsActive("e1") || list.IsActive("e2") {
		t.Errorf("IsActive mismatch for %+v", list)
	}
}
