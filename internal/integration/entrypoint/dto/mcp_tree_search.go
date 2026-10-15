package dto

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const mcpTreeSearchDefaultLimit = 50

type MCPTreeSearch struct {
	Query string `json:"query" description:"Text to search names, URLs and descriptions for." required_for:"search"`
	Limit int    `json:"limit" description:"Most results to return, 0 to 100."`
}

func (m *MCPTreeSearch) HasQuery() bool {
	return strings.TrimSpace(m.Query) != ""
}

func (m *MCPTreeSearch) HasValidLimit() bool {
	return m.Limit >= 0 && m.Limit <= entity.AIPageSize
}

func (m *MCPTreeSearch) ResultLimit() int {
	if m.Limit == 0 {
		return mcpTreeSearchDefaultLimit
	}
	return m.Limit
}

func (m *MCPTreeSearch) SearchQuery() string {
	return m.Query
}

func (m *MCPTreeSearch) SearchTree(localId string) SearchTree {
	return NewSearchTree(localId, m.Query)
}
