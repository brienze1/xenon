package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

var rawContentTypes = map[enums.RawLanguage]string{
	rawlanguage.Json:       "application/json",
	rawlanguage.Xml:        "application/xml",
	rawlanguage.Text:       "text/plain",
	rawlanguage.Html:       "text/html",
	rawlanguage.JavaScript: "application/javascript",
}

type BodyRaw struct {
	Language enums.RawLanguage
	Content  string
}

func NewBodyRaw(language, content string) BodyRaw {
	raw := BodyRaw{Language: rawlanguage.Json, Content: content}
	if language != "" {
		raw.Language = rawlanguage.GetRawLanguage(language)
	}
	return raw
}

func (b BodyRaw) ContentType() string {
	if contentType, found := rawContentTypes[b.Language]; found {
		return contentType
	}
	return rawContentTypes[rawlanguage.Text]
}
