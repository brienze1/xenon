package rightpanelview

import "testing"

func TestGetRightPanelView(t *testing.T) {
	for _, raw := range []string{"info", "INFO", " info "} {
		if got := GetRightPanelView(raw); got != Info {
			t.Errorf("GetRightPanelView(%q) = %q, want info", raw, got)
		}
	}
	for _, raw := range []string{"variables", "VARIABLES", " variables "} {
		if got := GetRightPanelView(raw); got != Variables {
			t.Errorf("GetRightPanelView(%q) = %q, want variables", raw, got)
		}
	}
	for _, raw := range []string{"code", "CODE", " code "} {
		if got := GetRightPanelView(raw); got != Code {
			t.Errorf("GetRightPanelView(%q) = %q, want code", raw, got)
		}
	}
	if got := GetRightPanelView("not-a-value"); got != Unknown {
		t.Errorf("GetRightPanelView(unknown) = %q, want unknown", got)
	}
}

func TestRightPanelViewEqualsAndString(t *testing.T) {
	if !Info.Equals("INFO") || Info.Equals("infox") || Info.String() != "info" {
		t.Errorf("Equals/String misbehave")
	}
}
