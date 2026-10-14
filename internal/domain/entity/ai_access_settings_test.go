package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/aiaccessstate"
)

func TestAIAccessSettingsPortRules(t *testing.T) {
	override := 50000
	cases := []struct {
		name      string
		settings  AIAccessSettings
		override  *int
		valid     bool
		effective int
	}{
		{name: "default", settings: DefaultAIAccessSettings(), valid: true, effective: DefaultAIPort},
		{name: "override replaces the default", settings: DefaultAIAccessSettings(), override: &override, valid: true, effective: override},
		{name: "saved port wins", settings: AIAccessSettings{Port: 48000}, override: &override, valid: true, effective: 48000},
		{name: "lowest", settings: AIAccessSettings{Port: MinAIPort}, valid: true, effective: MinAIPort},
		{name: "highest", settings: AIAccessSettings{Port: MaxAIPort}, valid: true, effective: MaxAIPort},
		{name: "too low", settings: AIAccessSettings{Port: 1023}, valid: false, effective: 1023},
		{name: "too high", settings: AIAccessSettings{Port: 65536}, valid: false, effective: 65536},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := test.settings.HasValidPort(); got != test.valid {
				t.Errorf("HasValidPort() = %v, want %v", got, test.valid)
			}
			if got := test.settings.EffectivePort(test.override); got != test.effective {
				t.Errorf("EffectivePort() = %d, want %d", got, test.effective)
			}
		})
	}
	if !DefaultAIAccessSettings().Enabled || DefaultAIAccessSettings().SkipApprovals {
		t.Error("AI access is on by default, approvals are not skipped")
	}
}

func TestAppSettingsMergeKeepsOmittedFields(t *testing.T) {
	timeout, cap, enabled, port := 2500, 250, false, 48000
	current := DefaultAppSettings()
	current.AIAccess.SkipApprovals = true
	merged := current.Merged(AppSettingsPatch{ScriptTimeoutMs: &timeout})
	if merged.ScriptTimeoutMs != 2500 || merged.HistoryUserCap != DefaultHistoryCap || !merged.AIAccess.SkipApprovals {
		t.Fatalf("partial merge = %+v", merged)
	}
	merged = current.Merged(AppSettingsPatch{HistoryAICap: &cap, HistoryUserCap: &cap, AIAccess: &AIAccessPatch{Enabled: &enabled, Port: &port}})
	if merged.HistoryAICap != 250 || merged.HistoryUserCap != 250 || merged.AIAccess.Enabled || merged.AIAccess.Port != port || !merged.AIAccess.SkipApprovals {
		t.Fatalf("full merge = %+v", merged)
	}
	if caps := merged.HistoryCaps(); caps != (HistoryCaps{User: 250, AI: 250}) {
		t.Errorf("HistoryCaps() = %+v", caps)
	}
	if caps := (AppSettings{}).HistoryCaps(); caps != DefaultHistoryCaps {
		t.Errorf("zero caps = %+v", caps)
	}
	if logged := merged.LogValue().(map[string]any); logged["history_ai_cap"] != 250 {
		t.Errorf("LogValue() = %v", logged)
	}
	if logged := (AppSettingsPatch{}).LogValue().(map[string]any); len(logged) != 4 {
		t.Errorf("patch LogValue() = %v", logged)
	}
	for limit, valid := range map[int]bool{99: false, 100: true, 20000: true, 20001: false} {
		if IsValidHistoryCap(limit) != valid {
			t.Errorf("IsValidHistoryCap(%d) = %v", limit, !valid)
		}
	}
}

func TestAIAccessStatusTransitions(t *testing.T) {
	running := RunningAIAccessStatus(47900, 0).WithClients(2)
	if !running.IsRunning() || running.Address() != "http://127.0.0.1:47900/mcp" || running.Clients != 2 {
		t.Fatalf("running = %+v", running)
	}
	off := OffAIAccessStatus(47900).WithClients(3)
	if off.State != aiaccessstate.Off || off.Address() != "" || off.Clients != 0 || off.URL() != "http://127.0.0.1:47900/mcp" {
		t.Fatalf("off = %+v", off)
	}
	busy := PortInUseAIAccessStatus(47900)
	if busy.State != aiaccessstate.Error || busy.Reason != "port 47900 is in use" || busy.Code != PortInUseCode || AIAccessListenAddress(47900) != "127.0.0.1:47900" {
		t.Fatalf("busy = %+v", busy)
	}
	if logged := busy.LogValue().(map[string]any); logged["reason"] != busy.Reason {
		t.Errorf("LogValue() = %v", logged)
	}
	if logged := (AIAccessSettings{Port: 1}).LogValue().(map[string]any); logged["port"] != 1 {
		t.Errorf("settings LogValue() = %v", logged)
	}
}
