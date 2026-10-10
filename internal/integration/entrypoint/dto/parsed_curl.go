package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ParsedCurl struct {
	Notices []string `json:"notices"`
	Request Request  `json:"request"`
}

func NewParsedCurl(parsed entity.ParsedCurl) ParsedCurl {
	return ParsedCurl{Request: NewRequest(parsed.Request), Notices: append([]string{}, parsed.Notices...)}
}

func (p ParsedCurl) LogValue() any {
	return map[string]any{"request": p.Request.LogValue(), "notices": p.Notices}
}
