package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/aiactivityoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
)

func TestAIToolRules(t *testing.T) {
	cases := []struct {
		name        string
		tool        AITool
		argument    string
		hasActions  bool
		defaulted   string
		readOnly    bool
		destructive bool
	}{
		{name: "read tool", tool: AITool{Actions: []string{"list", "get"}, Effect: aitooleffect.ReadOnly}, argument: "action", hasActions: true, readOnly: true},
		{name: "single action", tool: AITool{Actions: []string{"apply"}, Effect: aitooleffect.Destructive}, argument: "action", hasActions: true, defaulted: "apply", destructive: true},
		{name: "selector", tool: AITool{Selector: "target", Actions: []string{"item", "cookie"}, Effect: aitooleffect.Destructive}, argument: "target", hasActions: true, destructive: true},
		{name: "no actions", tool: AITool{Effect: aitooleffect.Write}, argument: "action"},
	}
	for _, tc := range cases {
		tool := tc.tool
		if tool.ArgumentName() != tc.argument || tool.HasActions() != tc.hasActions || tool.DefaultAction() != tc.defaulted || tool.IsReadOnly() != tc.readOnly || tool.IsDestructive() != tc.destructive {
			t.Errorf("%s: argument %q actions %v default %q readOnly %v destructive %v", tc.name, tool.ArgumentName(), tool.HasActions(), tool.DefaultAction(), tool.IsReadOnly(), tool.IsDestructive())
		}
	}
	tool := AITool{Actions: []string{"list", "get"}}
	for action, want := range map[string]bool{"list": true, "get": true, "LIST": false, "": false, "delete": false} {
		if tool.Allows(action) != want {
			t.Errorf("Allows(%q) = %v", action, !want)
		}
	}
}

func TestAIActivityOutcomeOf(t *testing.T) {
	cases := map[string]string{
		"":          aiactivityoutcome.Ok.String(),
		"MCP-03403": aiactivityoutcome.Denied.String(),
		"MCP-04408": aiactivityoutcome.TimedOut.String(),
		"MCP-07400": aiactivityoutcome.Error.String(),
		"UNK-02500": aiactivityoutcome.Error.String(),
	}
	for code, want := range cases {
		if got := AIActivityOutcomeOf(code); got.String() != want {
			t.Errorf("AIActivityOutcomeOf(%q) = %s, want %s", code, got, want)
		}
	}
}
