package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type historyListArguments struct {
	Source string `json:"source"`
	Query  string `json:"query"`
	Cursor string `json:"cursor"`
}

type historyGetArguments struct {
	Id string `json:"id"`
}

type historyLine struct {
	Status  *int   `json:"status,omitempty"`
	Id      string `json:"id"`
	At      string `json:"at"`
	Kind    string `json:"kind"`
	Source  string `json:"source"`
	Client  string `json:"client,omitempty"`
	Method  string `json:"method,omitempty"`
	Url     string `json:"url,omitempty"`
	Tool    string `json:"tool,omitempty"`
	Action  string `json:"action,omitempty"`
	Target  string `json:"target,omitempty"`
	Outcome string `json:"outcome,omitempty"`
	Ms      int64  `json:"ms,omitempty"`
}

type historySendView struct {
	Status  *int              `json:"status,omitempty"`
	Request historyRequestRow `json:"request"`
	Error   string            `json:"error,omitempty"`
	Ms      int64             `json:"ms"`
	Size    int64             `json:"size"`
}

type historyRequestRow struct {
	Auth    *dto.Auth      `json:"auth,omitempty"`
	Body    *dto.Body      `json:"body,omitempty"`
	Name    string         `json:"name,omitempty"`
	Method  string         `json:"method"`
	Url     string         `json:"url"`
	Params  []dto.KeyValue `json:"params,omitempty"`
	Headers []dto.KeyValue `json:"headers,omitempty"`
}

type historyActionView struct {
	Target  *dto.AIActivityTarget `json:"target,omitempty"`
	Tool    string                `json:"tool"`
	Action  string                `json:"action,omitempty"`
	Outcome string                `json:"outcome"`
	Code    string                `json:"code,omitempty"`
	Args    string                `json:"args,omitempty"`
	Ms      int64                 `json:"ms"`
}

func newHistoryLine(entry dto.HistoryEntry) historyLine {
	line := historyLine{Id: entry.Id, At: entry.SentAt, Kind: entry.Kind, Source: entry.Source, Client: entry.Client, Ms: entry.TotalMs}
	if entry.Activity == nil {
		line.Method, line.Url, line.Status, line.Outcome = entry.Method, entry.Url, entry.StatusCode, entry.Outcome
		return line
	}
	line.Tool, line.Action, line.Outcome = entry.Activity.Tool, entry.Activity.Action, entry.Activity.Outcome
	if entry.Activity.Target != nil {
		line.Target = entry.Activity.Target.Name
	}
	return line
}

func newHistoryRequestRow(request dto.Request) historyRequestRow {
	row := historyRequestRow{Name: request.Name, Method: request.Method, Url: request.Url, Params: request.Params, Headers: request.Headers}
	if authType := authtype.GetAuthType(request.Auth.Type); authType != authtype.None && authType != authtype.Inherit {
		row.Auth = &request.Auth
	}
	if bodymode.GetBodyMode(request.Body.Mode) != bodymode.None {
		row.Body = &request.Body
	}
	return row
}

func (t *tools) history(ctx context.Context, call Call) (Result, error) {
	if call.Action == "get" {
		return t.historyGet(ctx, call)
	}
	return t.historyList(ctx, call)
}

func (t *tools) historyList(ctx context.Context, call Call) (Result, error) {
	var arguments historyListArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	entries, err := t.controllers.History.List(ctx, dto.ListHistory{LocalId: call.LocalId, Source: arguments.Source, Query: arguments.Query, BeforeId: arguments.Cursor, Limit: PageSize + 1})
	if err != nil {
		return Result{}, err
	}
	page := List[historyLine]{Items: make([]historyLine, 0, min(len(entries), PageSize))}
	for index, entry := range entries {
		if index == PageSize {
			page.Next = entries[index-1].Id
			break
		}
		page.Items = append(page.Items, newHistoryLine(entry))
	}
	return Result{Payload: page}, nil
}

func (t *tools) historyGet(ctx context.Context, call Call) (Result, error) {
	var arguments historyGetArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	detail, err := t.controllers.History.Get(ctx, dto.HistoryReference{LocalId: call.LocalId, Id: arguments.Id})
	if err != nil {
		return Result{}, err
	}
	target := dto.AIActivityTarget{Id: detail.Entry.Id}
	if detail.Activity != nil {
		activity := detail.Activity
		return Result{Payload: historyActionView{Tool: activity.Tool, Action: activity.Action, Target: activity.Target, Outcome: activity.Outcome, Code: activity.Code, Ms: activity.DurationMs, Args: activity.Args}, Target: target}, nil
	}
	send := historySendView{Request: newHistoryRequestRow(detail.Request), Status: detail.Entry.StatusCode, Ms: detail.Entry.TotalMs, Size: detail.Entry.SizeBytes, Error: detail.Entry.ErrorKind}
	return Result{Payload: send, Target: target}, nil
}
