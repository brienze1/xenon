package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type scriptCheckArguments struct {
	Text string `json:"text"`
}

type scriptMemberLine struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type scriptReferenceLines struct {
	Members []scriptMemberLine `json:"members"`
}

type scriptSnippetLine struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func (t *tools) scripts(ctx context.Context, call Call) (Result, error) {
	switch call.Action {
	case "reference":
		return t.scriptReference(ctx)
	case "snippets":
		return t.scriptSnippets(ctx)
	default:
		return t.checkScript(ctx, call)
	}
}

func (t *tools) checkScript(ctx context.Context, call Call) (Result, error) {
	var arguments scriptCheckArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	check, err := t.controllers.ScriptTools.Check(ctx, dto.CheckScript{Text: arguments.Text})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: check}, nil
}

func (t *tools) scriptReference(ctx context.Context) (Result, error) {
	reference, err := t.controllers.ScriptTools.Reference(ctx)
	if err != nil {
		return Result{}, err
	}
	members := make([]scriptMemberLine, 0, len(reference.Members))
	for _, member := range reference.Members {
		members = append(members, scriptMemberLine{Path: member.Path, Kind: member.Kind})
	}
	return Result{Payload: scriptReferenceLines{Members: members}}, nil
}

func (t *tools) scriptSnippets(ctx context.Context) (Result, error) {
	snippets, err := t.controllers.ScriptTools.Snippets(ctx)
	if err != nil {
		return Result{}, err
	}
	lines := make([]scriptSnippetLine, 0, len(snippets))
	for _, snippet := range snippets {
		lines = append(lines, scriptSnippetLine{Id: snippet.Id, Name: snippet.Title, Code: snippet.Text})
	}
	return Result{Payload: List[scriptSnippetLine]{Items: lines}}, nil
}
