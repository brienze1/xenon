package dto

import "time"

type MCPTreeEditToken struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Type      string     `json:"type"`
	Scopes    string     `json:"scopes,omitempty"`
	State     string     `json:"state"`
	Use       string     `json:"use,omitempty"`
}

func NewMCPTreeEditToken(token TokenInfo) MCPTreeEditToken {
	return MCPTreeEditToken{ExpiresAt: token.ExpiresAt, Type: token.TokenType, Scopes: token.Scope, State: mcpTokenState(token).String(), Use: token.UseToken}
}
