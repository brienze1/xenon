package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/consoleentrykind"
	"github.com/brienze1/xenon/internal/domain/enums/consolefilter"
)

const (
	mcpConsoleReadLimit = 5000
	mcpConsoleAnyType   = ""
)

var mcpConsoleFilters = map[string]string{
	mcpConsoleAnyType:                 consolefilter.All.String(),
	consoleentrykind.Request.String(): consolefilter.Requests.String(),
	consoleentrykind.Log.String():     consolefilter.Logs.String(),
	consoleentrykind.Error.String():   consolefilter.Errors.String(),
}

type MCPConsoleList struct {
	Type   string `json:"type"`
	Query  string `json:"query"`
	Cursor string `json:"cursor"`
}

// ListConsole reads up to the whole console buffer; false means the type is not request, log or error.
func (m *MCPConsoleList) ListConsole() (ListConsole, bool) {
	filter, known := mcpConsoleFilters[m.Type]
	if !known {
		return ListConsole{}, false
	}
	return NewListConsole(filter, m.Query, mcpConsoleReadLimit), true
}

// Page pages the entries from the cursor; false means an unknown cursor.
func (m *MCPConsoleList) Page(entries []entity.ConsoleEntry) (entity.AIPage[entity.ConsoleEntry], bool) {
	return entity.NewAIPage(entries, m.Cursor)
}
