package dto

type MCPServerConfig struct {
	Type    string   `json:"type,omitempty"`
	Url     string   `json:"url,omitempty"`
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}

func NewHTTPMCPServerConfig(url string) MCPServerConfig {
	return MCPServerConfig{Type: httpTransportType, Url: url}
}

func NewStdioMCPServerConfig(executable string) MCPServerConfig {
	return MCPServerConfig{Command: executable, Args: []string{stdioBridgeCommand}}
}
