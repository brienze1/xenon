package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const consoleReadLimit = 5000

var consoleFilters = map[string]string{"": "all", "request": "requests", "log": "logs", "error": "errors"}

type consoleListArguments struct {
	Type   string `json:"type"`
	Query  string `json:"query"`
	Cursor string `json:"cursor"`
}

type consoleGetArguments struct {
	Id string `json:"id"`
}

type consoleLine struct {
	Id      string `json:"id"`
	At      string `json:"at"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

func (t *tools) console(ctx context.Context, call Call) (Result, error) {
	if call.Action == "get" {
		return t.consoleGet(ctx, call)
	}
	return t.consoleList(ctx, call)
}

func (t *tools) consoleList(ctx context.Context, call Call) (Result, error) {
	var arguments consoleListArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	filter, known := consoleFilters[arguments.Type]
	if !known {
		return Result{}, invalidArguments(call.Tool, "type must be request, log or error")
	}
	entries, err := t.controllers.Console.List(ctx, dto.ListConsole{Filter: filter, Search: arguments.Query, Limit: consoleReadLimit})
	if err != nil {
		return Result{}, err
	}
	lines := make([]consoleLine, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, consoleLine{Id: entry.Id, At: entry.CreatedAt, Type: entry.Kind, Summary: entry.Summary})
	}
	page, err := Paginate(call.Tool, lines, arguments.Cursor)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: page}, nil
}

func (t *tools) consoleGet(ctx context.Context, call Call) (Result, error) {
	var arguments consoleGetArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	entry, err := t.controllers.Console.Get(ctx, dto.ConsoleEntryReference{Id: arguments.Id})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: entry, Target: dto.AIActivityTarget{Id: entry.Id}}, nil
}
