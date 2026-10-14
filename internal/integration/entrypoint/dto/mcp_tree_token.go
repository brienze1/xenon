package dto

import (
	"strings"
	"time"

	"github.com/brienze1/xenon/internal/integration/entrypoint/enums"
	"github.com/brienze1/xenon/internal/integration/entrypoint/enums/mcptokenstate"
)

type MCPTreeToken struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Type      string     `json:"type"`
	State     string     `json:"state"`
	Scopes    []string   `json:"scopes,omitempty"`
}

func NewMCPTreeToken(token TokenInfo) MCPTreeToken {
	return MCPTreeToken{ExpiresAt: token.ExpiresAt, Type: token.TokenType, State: mcpTokenState(token).String(), Scopes: strings.Fields(token.Scope)}
}

func mcpTokenState(token TokenInfo) enums.MCPTokenState {
	if token.Expired {
		return mcptokenstate.Expired
	}
	return mcptokenstate.Valid
}
