package entity

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aiaccessstate"
)

const (
	aiAccessHost      = "127.0.0.1"
	aiAccessPath      = "/mcp"
	PortInUseCode     = "MCP-01409"
	portInUsePrefix   = "port "
	portInUseSuffix   = " is in use"
	aiAccessURLScheme = "http://"
)

type AIAccessStatus struct {
	State   enums.AIAccessState
	Reason  string
	Code    string
	Port    int
	Clients int
}

func OffAIAccessStatus(port int) AIAccessStatus {
	return AIAccessStatus{State: aiaccessstate.Off, Port: port}
}

func RunningAIAccessStatus(port, clients int) AIAccessStatus {
	return AIAccessStatus{State: aiaccessstate.Running, Port: port, Clients: clients}
}

func PortInUseAIAccessStatus(port int) AIAccessStatus {
	return AIAccessStatus{State: aiaccessstate.Error, Port: port, Code: PortInUseCode, Reason: portInUsePrefix + strconv.Itoa(port) + portInUseSuffix}
}

func AIAccessListenAddress(port int) string {
	return aiAccessHost + ":" + strconv.Itoa(port)
}

func AIAccessURL(port int) string {
	return aiAccessURLScheme + AIAccessListenAddress(port) + aiAccessPath
}

func (a AIAccessStatus) IsRunning() bool {
	return a.State == aiaccessstate.Running
}

func (a AIAccessStatus) Address() string {
	if !a.IsRunning() {
		return ""
	}
	return AIAccessURL(a.Port)
}

func (a AIAccessStatus) URL() string {
	return AIAccessURL(a.Port)
}

func (a AIAccessStatus) WithClients(clients int) AIAccessStatus {
	if a.IsRunning() {
		a.Clients = clients
	}
	return a
}

func (a AIAccessStatus) LogValue() any {
	return map[string]any{"state": a.State, "port": a.Port, "reason": a.Reason, "clients": a.Clients}
}
