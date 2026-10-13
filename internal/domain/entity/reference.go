package entity

import (
	"regexp"
	"strings"
	"unicode/utf16"
)

const dynamicVariablePrefix = "$"

var variableReference = regexp.MustCompile(`\{\{([^{}]*?)\}\}`)

type Reference struct {
	Name  string
	Field string
	Start int
	End   int
}

func TokenizeReferences(field, text string) []Reference {
	matches := variableReference.FindAllStringSubmatchIndex(text, -1)
	references := make([]Reference, 0, len(matches))
	for _, match := range matches {
		references = append(references, Reference{Name: text[match[2]:match[3]], Field: field, Start: match[0], End: match[1]})
	}
	return references
}

func (r Reference) Placeholder() string {
	return "{{" + r.Name + "}}"
}

func (r Reference) IsDynamic() bool {
	return strings.HasPrefix(r.Name, dynamicVariablePrefix)
}

// WithUtf16Offsets converts the byte offsets into the text to UTF-16 code units.
func (r Reference) WithUtf16Offsets(text string) Reference {
	start := utf16Length(text[:r.Start])
	r.End = start + utf16Length(text[r.Start:r.End])
	r.Start = start
	return r
}

func utf16Length(text string) int {
	return len(utf16.Encode([]rune(text)))
}
