package dto

import "github.com/brienze1/xenon/internal/domain/entity"

func (s SendResult) InlineText() (string, bool) {
	if s.InlineBody == nil {
		return "", false
	}
	return *s.InlineBody, true
}

func (s SendResult) HasAITextBody() bool {
	return s.Ok && !entity.IsAIBinaryBody(s.IsBinary, s.ContentType)
}
