package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type snippetArguments struct {
	Request        *dto.Request `json:"request"`
	RequestId      string       `json:"request_id"`
	Language       string       `json:"language"`
	IncludeSecrets bool         `json:"include_secrets"`
}

type snippetLanguageLine struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type snippetView struct {
	Code    string   `json:"code"`
	Notices []string `json:"notices,omitempty"`
}

func (t *tools) snippet(ctx context.Context, call Call) (Result, error) {
	if call.Action == "languages" {
		return t.snippetLanguages(ctx)
	}
	return t.generateSnippet(ctx, call)
}

func (t *tools) snippetLanguages(ctx context.Context) (Result, error) {
	languages, err := t.controllers.Snippet.Languages(ctx)
	if err != nil {
		return Result{}, err
	}
	lines := make([]snippetLanguageLine, 0, len(languages))
	for _, language := range languages {
		lines = append(lines, snippetLanguageLine{Id: language.Id, Name: language.Label})
	}
	return Result{Payload: List[snippetLanguageLine]{Items: lines}}, nil
}

func (t *tools) generateSnippet(ctx context.Context, call Call) (Result, error) {
	var arguments snippetArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	if arguments.IncludeSecrets {
		return Result{}, snippetSecretsRefused()
	}
	request, target, err := t.sendTarget(ctx, call, sendArguments{Request: arguments.Request, RequestId: arguments.RequestId})
	if err != nil {
		return Result{Target: target}, err
	}
	snippet, err := t.controllers.Snippet.Generate(ctx, dto.GenerateSnippet{LocalId: call.LocalId, TabId: call.TabId(), Language: arguments.Language, Request: request})
	if err != nil {
		return Result{Target: target}, err
	}
	return Result{Payload: snippetView{Code: snippet.Text, Notices: snippet.Notices}, Target: target}, nil
}
