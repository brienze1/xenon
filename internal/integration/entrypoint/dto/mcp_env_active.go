package dto

type MCPEnvActive struct {
	ActiveId string `json:"active_id"`
}

func NewMCPEnvActive(activeId string) MCPEnvActive {
	return MCPEnvActive{ActiveId: activeId}
}
