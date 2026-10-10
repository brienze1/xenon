package entity

type ParsedCurl struct {
	Notices []string
	Request Request
}

func (p ParsedCurl) WithRequest(request Request) ParsedCurl {
	p.Request = request
	return p
}

func (p ParsedCurl) LogValue() any {
	return map[string]any{"request": p.Request.LogValue(), "notices": p.Notices}
}
