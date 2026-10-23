package entity

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/consoleentrykind"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
)

const (
	consoleSummaryLimit     = 200
	consoleBodyKeyPrefix    = "console:"
	consoleErrorStatusFloor = 400
)

type ConsoleEntry struct {
	CreatedAt time.Time
	Request   *EffectiveRequest
	Response  *Response
	Script    *ScriptRef
	Kind      enums.ConsoleEntryKind
	Level     enums.ConsoleLevel
	ID        string
	SendID    string
	Summary   string
	Arguments []string
	Values    []any
	Line      int
	Column    int
	Masked    bool
	Nested    bool
}

func NewConsoleLogEntry(id, sendId string, script ScriptRef, log ScriptLog, createdAt time.Time) ConsoleEntry {
	return ConsoleEntry{
		ID:        id,
		SendID:    sendId,
		Kind:      consoleentrykind.Log,
		Level:     log.Level,
		Script:    &script,
		Summary:   ConsoleSummary(log.Text()),
		Arguments: append([]string{}, log.Arguments...),
		Values:    append([]any{}, log.Values...),
		CreatedAt: createdAt,
	}
}

func NewConsoleWarningEntry(id, sendId, message string, createdAt time.Time) ConsoleEntry {
	return ConsoleEntry{
		ID:        id,
		SendID:    sendId,
		Kind:      consoleentrykind.Log,
		Level:     consolelevel.Warn,
		Summary:   ConsoleSummary(message),
		Arguments: []string{message},
		Values:    []any{},
		CreatedAt: createdAt,
	}
}

func NewConsoleErrorEntry(id, sendId string, scriptError ScriptError, createdAt time.Time) ConsoleEntry {
	script := scriptError.Script
	return ConsoleEntry{
		ID:        id,
		SendID:    sendId,
		Kind:      consoleentrykind.Error,
		Level:     consolelevel.Error,
		Script:    &script,
		Summary:   ConsoleSummary(scriptError.Message),
		Arguments: []string{scriptError.Message},
		Line:      scriptError.Line,
		Column:    scriptError.Column,
		CreatedAt: createdAt,
	}
}

func NewConsoleTestFailureEntry(id, sendId string, test TestResult, createdAt time.Time) ConsoleEntry {
	script := test.Script
	return ConsoleEntry{
		ID:        id,
		SendID:    sendId,
		Kind:      consoleentrykind.TestFailure,
		Level:     consolelevel.Error,
		Script:    &script,
		Summary:   ConsoleSummary(test.Name + ": " + test.ErrorMessage),
		Arguments: []string{test.Name, test.ErrorMessage},
		CreatedAt: createdAt,
	}
}

// NewConsoleRequestEntry records a request that went out; script is the script that sent it (nil for a send's own request).
func NewConsoleRequestEntry(id, sendId string, script *ScriptRef, nested bool, request EffectiveRequest, response Response, createdAt time.Time) ConsoleEntry {
	entry := ConsoleEntry{
		ID:        id,
		SendID:    sendId,
		Kind:      consoleentrykind.Request,
		Level:     consolelevel.Info,
		Script:    script,
		Summary:   ConsoleSummary(request.Method + " " + request.URL),
		Arguments: []string{},
		Request:   &request,
		Response:  &response,
		Nested:    nested,
		CreatedAt: createdAt,
	}
	if !response.IsOk() || response.StatusCode >= consoleErrorStatusFloor {
		entry.Level = consolelevel.Error
	}
	return entry
}

// WithSecretsMasked replaces secret values everywhere in the entry; mask reports whether it replaced anything.
func (c ConsoleEntry) WithSecretsMasked(mask func(text string) (string, bool)) ConsoleEntry {
	masker := secretMasker{mask: mask}
	c.Summary = masker.text(c.Summary)
	c.Arguments = masker.texts(c.Arguments)
	c.Values = masker.values(c.Values)
	if c.Request != nil {
		request := masker.request(*c.Request)
		c.Request = &request
	}
	if c.Response != nil {
		response := masker.response(*c.Response)
		c.Response = &response
	}
	c.Masked = c.Masked || masker.masked
	return c
}

func ConsoleSummary(text string) string {
	if utf8.RuneCountInString(text) <= consoleSummaryLimit {
		return text
	}
	return string([]rune(text)[:consoleSummaryLimit])
}

func ConsoleBodyKey(id string) string {
	return consoleBodyKeyPrefix + id
}

func (c ConsoleEntry) BodyKey() string {
	return ConsoleBodyKey(c.ID)
}

// ServesBodyByRoute reports a response body too large or too binary to travel inline with the entry.
func (c ConsoleEntry) ServesBodyByRoute() bool {
	return c.Response != nil && c.Response.IsOk() && !c.Response.HasInlineBody()
}

func (c ConsoleEntry) WithoutRoutedBody() ConsoleEntry {
	if c.ServesBodyByRoute() {
		response := c.Response.WithoutBody()
		c.Response = &response
	}
	return c
}

func (c ConsoleEntry) WithoutResponseBody() ConsoleEntry {
	if c.Response != nil {
		response := c.Response.WithoutBody()
		c.Response = &response
	}
	return c
}

func (c ConsoleEntry) IsRequest() bool {
	return c.Kind == consoleentrykind.Request
}

func (c ConsoleEntry) IsLog() bool {
	return c.Kind == consoleentrykind.Log || c.Kind == consoleentrykind.Info
}

func (c ConsoleEntry) IsError() bool {
	return c.Kind == consoleentrykind.Error || c.Kind == consoleentrykind.TestFailure || c.Level == consolelevel.Error
}

// Contains reports whether the entry's text, URL or script name contains search, ignoring case.
func (c ConsoleEntry) Contains(search string) bool {
	search = strings.ToLower(search)
	texts := append([]string{c.Summary}, c.Arguments...)
	if c.Request != nil {
		texts = append(texts, c.Request.URL)
	}
	if c.Script != nil {
		texts = append(texts, c.Script.OwnerName)
	}
	for _, text := range texts {
		if searchable(text, search) {
			return true
		}
	}
	return false
}
