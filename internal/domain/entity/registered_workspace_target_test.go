package entity

import "testing"

func TestRegisteredWorkspaceTargets(t *testing.T) {
	workspace := RegisteredWorkspace{LocalID: "ws-1", DisplayNameCache: "Payments", Path: "/tmp/payments"}
	if target := workspace.Target(); target != (AIActivityTarget{ID: "ws-1", Name: "Payments", Path: "/tmp/payments"}) {
		t.Errorf("Target() = %+v", target)
	}
	if target := workspace.NamedTarget(); target != (AIActivityTarget{ID: "ws-1", Name: "Payments"}) {
		t.Errorf("NamedTarget() = %+v", target)
	}
}
