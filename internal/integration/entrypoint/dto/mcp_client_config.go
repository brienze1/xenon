package dto

import "encoding/json"

type MCPClientConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

func NewMCPClientConfig(server MCPServerConfig) MCPClientConfig {
	return MCPClientConfig{MCPServers: map[string]MCPServerConfig{mcpServerName: server}}
}

func (m MCPClientConfig) JSON() string {
	config, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(config)
}
