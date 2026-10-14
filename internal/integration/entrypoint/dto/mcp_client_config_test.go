package dto

import "testing"

func TestMCPClientConfigJSON(t *testing.T) {
	cases := []struct {
		name   string
		server MCPServerConfig
		want   string
	}{
		{name: "http", server: NewHTTPMCPServerConfig("http://127.0.0.1:7777/mcp"), want: `{"mcpServers":{"xenon":{"type":"http","url":"http://127.0.0.1:7777/mcp"}}}`},
		{name: "stdio", server: NewStdioMCPServerConfig("/apps/xenon"), want: `{"mcpServers":{"xenon":{"command":"/apps/xenon","args":["mcp"]}}}`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := NewMCPClientConfig(test.server).JSON(); got != test.want {
				t.Errorf("JSON() = %s, want %s", got, test.want)
			}
		})
	}
}
