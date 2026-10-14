package entity

import "testing"

func TestAppSettingsSkipAIApprovalsOnlyWhenChosen(t *testing.T) {
	if DefaultAppSettings().SkipsAIApprovals() {
		t.Error("approvals must be asked by default")
	}
	settings := DefaultAppSettings()
	settings.AIAccess.SkipApprovals = true
	if !settings.SkipsAIApprovals() {
		t.Error("approvals must be skipped when the user chose to")
	}
}
