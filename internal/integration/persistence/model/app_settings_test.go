package model

import (
	"encoding/json"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aiactivityoutcome"
)

func TestAppSettingsFromAnOldAppJsonReadAsDefaults(t *testing.T) {
	var stored AppSettings
	if err := json.Unmarshal([]byte(`{"scriptTimeoutMs":3000}`), &stored); err != nil {
		t.Fatal(err)
	}
	got := stored.ToEntity()
	want := entity.DefaultAppSettings()
	want.ScriptTimeoutMs = 3000
	if got != want {
		t.Fatalf("ToEntity() = %+v, want %+v", got, want)
	}
}

func TestAppSettingsRoundTrip(t *testing.T) {
	settings := entity.AppSettings{
		ScriptTimeoutMs: 2500,
		HistoryUserCap:  10000,
		HistoryAICap:    100,
		AIAccess:        entity.AIAccessSettings{Enabled: false, Port: 48000, SkipApprovals: true},
	}
	data, err := json.Marshal(NewAppSettings(settings))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"aiAccess":{"enabled":false,"port":48000,"skipApprovals":true},"autoUpdate":false,"scriptTimeoutMs":2500,"historyUserCap":10000,"historyAiCap":100}` {
		t.Fatalf("json = %s", data)
	}
	var decoded AppSettings
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if got := decoded.ToEntity(); got != settings {
		t.Fatalf("round trip = %+v", got)
	}
}

func TestAppSettingsOutOfRangeFieldsFallBackOneByOne(t *testing.T) {
	stored := AppSettings{ScriptTimeoutMs: 5, HistoryUserCap: 50, HistoryAICap: 300, AIAccess: &AIAccessSettings{Port: 80}}
	got := stored.ToEntity()
	if got.ScriptTimeoutMs != entity.DefaultScriptTimeoutMs || got.HistoryUserCap != entity.DefaultHistoryCap || got.HistoryAICap != 300 || got.AIAccess.Port != 0 || !got.AIAccess.Enabled {
		t.Fatalf("ToEntity() = %+v", got)
	}
}

func TestHistoryEntryLinesWithoutSourceReadAsUserSends(t *testing.T) {
	var stored HistoryEntry
	if err := json.Unmarshal([]byte(`{"version":1,"id":"01","request":{"method":"GET","url":"https://old.test"},"result":{"status_code":200}}`), &stored); err != nil {
		t.Fatal(err)
	}
	entry := stored.ToEntity()
	if entry.Kind != "send" || entry.Source != "user" || entry.Request.URL != "https://old.test" || entry.Result.StatusCode != 200 {
		t.Fatalf("ToEntity() = %+v", entry)
	}
	action := entity.NewActionHistoryEntry("02", entry.SentAt, "", entity.AIActivity{Tool: "tree", Action: "list", Outcome: aiactivityoutcome.Ok, Target: entity.AIActivityTarget{ID: "1", Name: "A"}})
	data, _ := json.Marshal(NewHistoryEntry(action))
	var decoded HistoryEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	round := decoded.ToEntity()
	if !round.IsAction() || round.Client != entity.DefaultAIClientName || round.Activity.Target.Name != "A" || decoded.Request != nil {
		t.Fatalf("action round trip = %+v (%s)", round, data)
	}
}

func TestAppSettingsAutoUpdateIsOnWhenNeverStored(t *testing.T) {
	var stored AppSettings
	if err := json.Unmarshal([]byte(`{"scriptTimeoutMs":3000}`), &stored); err != nil {
		t.Fatal(err)
	}
	if !stored.ToEntity().UpdatesAutomatically() {
		t.Fatal("a machine that never stored the setting updates automatically")
	}
	if err := json.Unmarshal([]byte(`{"scriptTimeoutMs":3000,"autoUpdate":false}`), &stored); err != nil {
		t.Fatal(err)
	}
	if stored.ToEntity().UpdatesAutomatically() {
		t.Fatal("a stored false is kept")
	}
}
