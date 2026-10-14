package entity

import "testing"

func TestAIActivityRedactedArgs(t *testing.T) {
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
			activity := AIActivity{Tool: test.tool, Action: test.action, Args: test.args}
			if got := activity.RedactedArgs(); got != test.want {
				t.Errorf("RedactedArgs() = %s, want %s", got, test.want)
			}
		})
	}
}
