package dto

import (
	"strings"
)

const (
	detailProblemJoiner = ", "
)

type MCPToolErrorDetail struct {
	Field   string `json:"field"`
	Problem string `json:"problem"`
}

func NewMCPToolErrorDetail(field string, messages []string) MCPToolErrorDetail {
	return MCPToolErrorDetail{Field: field, Problem: strings.Join(messages, detailProblemJoiner)}
}
