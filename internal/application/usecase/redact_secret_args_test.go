package usecase

import (
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestRedactSecretArgs(t *testing.T) {
	cases := []struct {
		name   string
		tool   string
		action string
		args   string
		want   string
	}{
		{name: "env add_var value", tool: "env_edit", action: "add_var", args: `{"action":"add_var","key":"k","value":"s3cret"}`, want: `{"action":"add_var","key":"k","value":"•••••"}`},
		{name: "env set_current value", tool: "env_edit", action: "set_current", args: `{"action":"set_current","value":"s3cret"}`, want: `{"action":"set_current","value":"•••••"}`},
		{name: "plain value elsewhere", tool: "tree_edit", action: "update", args: `{"value":"kept"}`, want: `{"value":"kept"}`},
		{name: "value beside secret true", tool: "tree_edit", action: "save_settings", args: `{"variables":[{"key":"a","secret":true,"value":"s3cret"},{"key":"b","secret":false,"value":"plain"}]}`, want: `{"variables":[{"key":"a","secret":true,"value":"•••••"},{"key":"b","secret":false,"value":"plain"}]}`},
		{name: "secret_value anywhere", tool: "env_edit", action: "set_vars", args: `{"vars":[{"key":"a","secret_value":"s3cret"}]}`, want: `{"vars":[{"key":"a","secret_value":"•••••"}]}`},
		{name: "passphrase anywhere", tool: "import", action: "run", args: `{"passphrase":"p4ss","paths":["/a"]}`, want: `{"passphrase":"•••••","paths":["/a"]}`},
		{name: "html kept unescaped", tool: "tree", action: "search", args: `{"query":"<a&b>"}`, want: `{"query":"<a&b>"}`},
		{name: "not json", tool: "tree", action: "list", args: `not json`, want: `not json`},
		{name: "not an object", tool: "env_edit", action: "add_var", args: `["value"]`, want: `["value"]`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := RedactSecretArgs(test.tool, test.action, test.args); got != test.want {
				t.Errorf("RedactSecretArgs() = %s, want %s", got, test.want)
			}
		})
	}
}

func TestValidateAppSettingsReturnsTheFirstFailingCode(t *testing.T) {
	valid := validSettings()
	if err := ValidateAppSettings(valid); err != nil {
		t.Fatalf("valid settings: %v", err)
	}
	cases := map[string]func() error{
		"SCR-01400": func() error {
			s := validSettings()
			s.ScriptTimeoutMs, s.HistoryUserCap = 999, 99
			return ValidateAppSettings(s)
		},
		"SET-02400": func() error {
			s := validSettings()
			s.HistoryUserCap, s.HistoryAICap = 20001, 5
			return ValidateAppSettings(s)
		},
		"SET-03400": func() error { s := validSettings(); s.HistoryAICap = 99; return ValidateAppSettings(s) },
		"SET-04400": func() error { s := validSettings(); s.AIAccess.Port = 1023; return ValidateAppSettings(s) },
	}
	for code, validate := range cases {
		if err := validate(); err == nil || !strings.Contains(codeOf(err), code) {
			t.Errorf("want %s, got %v", code, err)
		}
	}
	if err := RejectResolvedApproval("01A"); !strings.Contains(codeOf(err), "MCP-05409") {
		t.Errorf("RejectResolvedApproval() = %v", err)
	}
}

func validSettings() entity.AppSettings {
	return entity.DefaultAppSettings()
}
