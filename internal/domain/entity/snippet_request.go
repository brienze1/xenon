package entity

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
)

type SnippetRequest struct {
	Method  string
	URL     string
	Body    Body
	Headers []KeyValue
	Notices []string
}

func NewSnippetRequest(request EffectiveRequest, workspaceRoot string, notices []string) SnippetRequest {
	headers := slices.Clone(request.Headers)
	contentType := request.Body.ContentType()
	if request.Body.HasContent() && !request.Body.IsMultipart() && !request.HasHeader(HeaderContentType) && !request.IsAutoHeaderDisabled(HeaderContentType) {
		headers = append(headers, NewKeyValue(HeaderContentType, contentType))
	}
	return SnippetRequest{Method: request.Method, URL: request.URL, Body: request.Body.RelativeTo(workspaceRoot), Headers: headers, Notices: slices.Clone(notices)}
}

func (s SnippetRequest) HasBody() bool {
	return s.Body.Mode != bodymode.None && s.Body.Mode != ""
}

func (s SnippetRequest) IsBodyEmpty() bool {
	switch s.Body.Mode {
	case bodymode.Raw:
		return s.Body.Raw.Content == ""
	case bodymode.UrlEncoded:
		return !slices.ContainsFunc(s.Body.UrlEncoded, KeyValue.IsEnabled)
	case bodymode.FormData:
		return !slices.ContainsFunc(s.Body.FormData, func(row FormDataRow) bool { return row.Enabled })
	case bodymode.GraphQl:
		return s.Body.GraphQl.Query == ""
	case bodymode.Binary:
		return s.Body.Binary.IsEmpty()
	default:
		return true
	}
}

func (s SnippetRequest) Header(name string) (string, bool) {
	for _, header := range s.Headers {
		if header.HasKey(name) {
			return header.Value, true
		}
	}
	return "", false
}

func (s SnippetRequest) LogValue() any {
	return map[string]any{"method": s.Method, "headers": len(s.Headers), "body": s.Body.LogValue(), "notices": s.Notices}
}
