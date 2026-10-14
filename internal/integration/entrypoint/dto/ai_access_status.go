package dto

import "github.com/brienze1/xenon/internal/domain/entity"

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

func NewAIAccessStatus(status entity.AIAccessStatus, executable string) AIAccessStatus {
	return AIAccessStatus{
		State:        status.State.String(),
		Address:      status.Address(),
		Reason:       status.Reason,
		Code:         status.Code,
		Clients:      status.Clients,
		StdioCommand: executable + " " + stdioBridgeCommand,
		HttpConfig:   NewMCPClientConfig(NewHTTPMCPServerConfig(status.URL())).JSON(),
		StdioConfig:  NewMCPClientConfig(NewStdioMCPServerConfig(executable)).JSON(),
	}
}
