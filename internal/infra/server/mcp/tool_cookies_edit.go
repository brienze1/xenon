package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type cookieUpsertArguments struct {
	Domain string `json:"domain"`
	Raw    string `json:"raw"`
}

type cookieAllowlistArguments struct {
	Domains []string `json:"domains"`
}

func (t *tools) cookiesEdit(ctx context.Context, call Call) (Result, error) {
	if call.Action == "allowlist" {
		return t.saveCookieAllowlist(ctx, call)
	}
	return t.upsertCookie(ctx, call)
}

func (t *tools) upsertCookie(ctx context.Context, call Call) (Result, error) {
	var arguments cookieUpsertArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	cookie, err := t.controllers.Cookie.Upsert(ctx, dto.UpsertCookie{LocalId: call.LocalId, Domain: arguments.Domain, Raw: arguments.Raw})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newCookieLine(*cookie), Target: dto.AIActivityTarget{Name: cookie.Name, Path: cookie.Domain + cookie.Path}, Changes: []dto.AIChange{{Kind: aichangekind.Cookies.String()}}}, nil
}

func (t *tools) saveCookieAllowlist(ctx context.Context, call Call) (Result, error) {
	var arguments cookieAllowlistArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	allowlist, err := t.controllers.Cookie.SaveAllowlist(ctx, dto.SaveCookieAllowlist{LocalId: call.LocalId, Domains: arguments.Domains})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: allowlist, Changes: []dto.AIChange{{Kind: aichangekind.Cookies.String()}}}, nil
}
