package mcp

import (
	"context"
	"strings"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type cookiesListArguments struct {
	Domain string `json:"domain"`
	Cursor string `json:"cursor"`
}

type cookieLine struct {
	Domain   string `json:"domain"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path"`
	Expires  string `json:"expires,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HttpOnly bool   `json:"http_only,omitempty"`
}

func newCookieLine(cookie dto.Cookie) cookieLine {
	return cookieLine{Domain: cookie.Domain, Name: cookie.Name, Value: cookie.Value, Path: cookie.Path, Expires: cookie.Expires, Secure: cookie.Secure, HttpOnly: cookie.HttpOnly}
}

func (t *tools) cookies(ctx context.Context, call Call) (Result, error) {
	if call.Action == "allowlist" {
		return t.cookieAllowlist(ctx, call)
	}
	return t.cookieList(ctx, call)
}

func (t *tools) cookieList(ctx context.Context, call Call) (Result, error) {
	var arguments cookiesListArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	domains, err := t.controllers.Cookie.List(ctx, dto.CookieJarReference{LocalId: call.LocalId})
	if err != nil {
		return Result{}, err
	}
	lines := make([]cookieLine, 0)
	for _, domain := range domains {
		if arguments.Domain != "" && !strings.EqualFold(domain.Domain, arguments.Domain) {
			continue
		}
		for _, cookie := range domain.Cookies {
			lines = append(lines, newCookieLine(cookie))
		}
	}
	page, err := Paginate(call.Tool, lines, arguments.Cursor)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: page}, nil
}

func (t *tools) cookieAllowlist(ctx context.Context, call Call) (Result, error) {
	allowlist, err := t.controllers.Cookie.GetAllowlist(ctx, dto.CookieJarReference{LocalId: call.LocalId})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: allowlist}, nil
}
