package dto

import (
	"fmt"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/importorigin"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
)

type GenerateSnippet struct {
	Options        map[string]any `json:"options"`
	LocalId        string         `json:"local_id" validate:"required,not_blank"`
	TabId          string         `json:"tab_id"`
	Language       string         `json:"language" validate:"required,oneof=curl js_fetch go_native python_requests"`
	Request        Request        `json:"request"`
	IncludeSecrets bool           `json:"include_secrets"`
}

func (g *GenerateSnippet) ToEntity() entity.SnippetInput {
	options := map[string]string{}
	for key, value := range g.Options {
		options[key] = fmt.Sprint(value)
	}
	request := g.Request.ToEntity()
	if unsupported := g.Request.Auth.Unsupported; unsupported != nil {
		request.Auth.Unsupported = entity.AuthUnsupported{Origin: importorigin.GetImportOrigin(unsupported.Source), Type: unsupported.Type}
	}
	return entity.SnippetInput{
		LocalID:        g.LocalId,
		TabID:          g.TabId,
		Request:        request,
		Options:        entity.SnippetOptions{Language: snippetlanguage.GetSnippetLanguage(g.Language), Options: options},
		IncludeSecrets: g.IncludeSecrets,
	}
}

func (g GenerateSnippet) LogValue() any {
	return map[string]any{"local_id": g.LocalId, "tab_id": g.TabId, "language": g.Language, "options": g.Options, "request": g.Request.LogValue(), "include_secrets": g.IncludeSecrets}
}
