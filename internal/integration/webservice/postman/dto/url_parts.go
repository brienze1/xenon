package dto

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	partPlaceholderStart = ""
	partPlaceholderEnd   = ""
)

var (
	urlVariable    = regexp.MustCompile(`{{[^{}]*}}`)
	urlPlaceholder = regexp.MustCompile(partPlaceholderStart + `(\d+)` + partPlaceholderEnd)
)

type UrlParts struct {
	Protocol string
	Port     string
	Hash     string
	Host     []string
	Path     []string
	HasPath  bool
}

// NewUrlParts splits a raw url the way Postman's url parser does, keeping {{variables}} whole.
func NewUrlParts(raw string) UrlParts {
	variables := urlVariable.FindAllString(raw, -1)
	index := 0
	text := urlVariable.ReplaceAllStringFunc(strings.TrimLeft(raw, " \t\r\n"), func(string) string {
		placeholder := partPlaceholderStart + strconv.Itoa(index) + partPlaceholderEnd
		index++
		return placeholder
	})
	parts := UrlParts{}
	text, parts.Hash, _ = strings.Cut(text, "#")
	text, _, _ = strings.Cut(text, "?")
	text = strings.ReplaceAll(text, `\`, "/")
	if protocol, rest, found := strings.Cut(text, "://"); found {
		parts.Protocol = protocol
		text = strings.TrimLeft(rest, "/")
		if strings.EqualFold(protocol, "file") {
			text = "/" + text
		}
	}
	if host, path, found := strings.Cut(text, "/"); found {
		parts.Path, parts.HasPath = strings.Split(path, "/"), true
		text = host
	}
	if at := strings.LastIndex(text, "@"); at >= 0 {
		text = text[at+1:]
	}
	if colon := strings.LastIndex(text, ":"); colon >= 0 && !strings.Contains(text[colon+1:], "]") {
		parts.Port = text[colon+1:]
		text = text[:colon]
	}
	if text != "" {
		parts.Host = strings.Split(text, ".")
	}
	return parts.restored(variables)
}

func (u UrlParts) restored(variables []string) UrlParts {
	restore := func(text string) string {
		return urlPlaceholder.ReplaceAllStringFunc(text, func(match string) string {
			position, _ := strconv.Atoi(urlPlaceholder.FindStringSubmatch(match)[1])
			return variables[position]
		})
	}
	u.Protocol, u.Port, u.Hash = restore(u.Protocol), restore(u.Port), restore(u.Hash)
	for index := range u.Host {
		u.Host[index] = restore(u.Host[index])
	}
	for index := range u.Path {
		u.Path[index] = restore(u.Path[index])
	}
	return u
}
