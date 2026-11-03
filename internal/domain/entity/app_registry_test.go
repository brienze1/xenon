package entity

import "testing"

func TestAppRegistryAddKeepsTheActiveWorkspace(t *testing.T) {
	registry := NewAppRegistry()
	registry.Register(RegisteredWorkspace{LocalID: "w1"})
	registry.Add(RegisteredWorkspace{LocalID: "w2"})
	if registry.ActiveLocalID != "w1" {
		t.Errorf("ActiveLocalID = %q, want w1", registry.ActiveLocalID)
	}
	if _, found := registry.Find("w2"); !found {
		t.Errorf("w2 not registered: %+v", registry.Workspaces)
	}
}

func TestAppRegistryWithActiveLeavesTheRegistryUntouched(t *testing.T) {
	registry := NewAppRegistry()
	registry.Register(RegisteredWorkspace{LocalID: "w1"})
	registry.Add(RegisteredWorkspace{LocalID: "w2"})
	shown := registry.WithActive("w2")
	if !shown.IsActive("w2") || shown.IsActive("w1") || !registry.IsActive("w1") {
		t.Errorf("shown = %q, registry = %q", shown.ActiveLocalID, registry.ActiveLocalID)
	}
}
