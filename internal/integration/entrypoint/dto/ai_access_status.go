package dto

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	mcpServerName      = "xenon"
	stdioBridgeCommand = "mcp"
	httpTransportType  = "http"
)

type AIAccessStatus struct {
	State        string `json:"state"`
	Address      string `json:"address,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Code         string `json:"code,omitempty"`
	StdioCommand string `json:"stdio_command"`
	HttpConfig   string `json:"http_config"`
	StdioConfig  string `json:"stdio_config"`
	Clients      int    `json:"clients"`
}

type mcpClientConfig struct {
	McpServers map[string]mcpServerConfig `json:"mcpServers"`
}

type mcpServerConfig struct {
	Type    string   `json:"type,omitempty"`
	Url     string   `json:"url,omitempty"`
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}

func NewAIAccessStatus(status entity.AIAccessStatus, executable string) AIAccessStatus {
	return AIAccessStatus{
		State:        status.State.String(),
		Address:      status.Address(),
		Reason:       status.Reason,
		Code:         status.Code,
		Clients:      status.Clients,
		StdioCommand: executable + " " + stdioBridgeCommand,
		HttpConfig:   clientConfig(mcpServerConfig{Type: httpTransportType, Url: status.URL()}),
		StdioConfig:  clientConfig(mcpServerConfig{Command: executable, Args: []string{stdioBridgeCommand}}),
	}
}

func clientConfig(server mcpServerConfig) string {
	config, _ := json.Marshal(mcpClientConfig{McpServers: map[string]mcpServerConfig{mcpServerName: server}})
	return string(config)
}
