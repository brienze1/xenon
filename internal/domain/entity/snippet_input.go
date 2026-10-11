package entity

type SnippetInput struct {
	Options        SnippetOptions
	LocalID        string
	TabID          string
	Request        Request
	IncludeSecrets bool
}

func (s SnippetInput) LogValue() any {
	return map[string]any{"localId": s.LocalID, "tabId": s.TabID, "request": s.Request.LogValue(), "language": s.Options.Language, "options": s.Options.Options, "includeSecrets": s.IncludeSecrets}
}
