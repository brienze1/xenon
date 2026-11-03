package entity

import "testing"

func TestResolveEnvironmentRef(t *testing.T) {
	environments := []Environment{{ID: "dev-id", Name: "dev"}, {ID: "prod-id", Name: "prod"}}
	tests := []struct {
		name      string
		ref       string
		wantId    string
		wantKnown bool
	}{
		{name: "active", ref: "", wantId: "", wantKnown: true},
		{name: "none", ref: NoEnvironmentID, wantId: NoEnvironmentID, wantKnown: true},
		{name: "by id", ref: "prod-id", wantId: "prod-id", wantKnown: true},
		{name: "by name", ref: "dev", wantId: "dev-id", wantKnown: true},
		{name: "unknown", ref: "staging", wantId: "", wantKnown: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, known := ResolveEnvironmentRef(environments, test.ref)
			if id != test.wantId || known != test.wantKnown {
				t.Errorf("ResolveEnvironmentRef(%q) = (%q, %v), want (%q, %v)", test.ref, id, known, test.wantId, test.wantKnown)
			}
		})
	}
}
