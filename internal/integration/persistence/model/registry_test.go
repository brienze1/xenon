package model

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestRegistryRoundTrip(t *testing.T) {
	registry := entity.NewAppRegistry()
	addedAt := time.Date(2026, 10, 9, 10, 30, 0, 123000000, time.FixedZone("BRT", -3*3600))
	registry.Register(entity.NewRegisteredWorkspace("01L", "/work/acme", "Acme", addedAt))
	model := NewRegistry(registry, Registry{Window: json.RawMessage(`{"width":1280}`), Vault: json.RawMessage(`{"mode":"keychain"}`), Settings: json.RawMessage(`{"scriptTimeoutMs":2000}`)})
	data, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Registry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Workspaces[0].AddedAt != "2026-10-09T13:30:00.123Z" || string(decoded.Window) != `{"width":1280}` || string(decoded.Vault) != `{"mode":"keychain"}` || string(decoded.Settings) != `{"scriptTimeoutMs":2000}` || decoded.Version != 1 {
		t.Fatalf("decoded = %+v", decoded)
	}
	got := decoded.ToEntity()
	want := registry
	want.Workspaces[0].AddedAt = addedAt.UTC()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ToEntity() = %+v, want %+v", got, want)
	}
}
