package snippets

import (
	"slices"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	booleanOption  = "boolean"
	positiveOption = "positiveInteger"
	enumOption     = "enum"

	optionMultiLine        = "multiLine"
	optionLongFormat       = "longFormat"
	optionLineContinuation = "lineContinuationCharacter"
	optionQuoteType        = "quoteType"
	optionTimeoutSeconds   = "requestTimeoutInSeconds"
	optionFollowRedirect   = "followRedirect"
	optionFollowMethod     = "followOriginalHttpMethod"
	optionTrimBody         = "trimRequestBody"
	optionSilent           = "silent"
	optionIndentCount      = "indentCount"
	optionIndentType       = "indentType"
	optionTimeout          = "requestTimeout"
	optionAsyncAwait       = "asyncAwaitEnabled"
)

var (
	indentCountOption    = entity.SnippetOptionSpec{ID: optionIndentCount, Label: "Set indentation count", Type: positiveOption, Default: "2"}
	indentTypeOption     = entity.SnippetOptionSpec{ID: optionIndentType, Label: "Set indentation type", Type: enumOption, Default: "Space", Values: []string{"Tab", "Space"}}
	timeoutOption        = entity.SnippetOptionSpec{ID: optionTimeout, Label: "Set request timeout (in milliseconds)", Type: positiveOption, Default: "0"}
	followRedirectOption = entity.SnippetOptionSpec{ID: optionFollowRedirect, Label: "Follow redirects", Type: booleanOption, Default: "true"}
	trimBodyOption       = entity.SnippetOptionSpec{ID: optionTrimBody, Label: "Trim request body fields", Type: booleanOption, Default: "false"}
)

type options map[string]string

func resolveOptions(specs []entity.SnippetOptionSpec, given map[string]string) options {
	resolved := options{}
	for _, spec := range specs {
		resolved[spec.ID] = spec.Default
		if value, found := given[spec.ID]; found && isValidOption(spec, value) {
			resolved[spec.ID] = value
		}
	}
	return resolved
}

func isValidOption(spec entity.SnippetOptionSpec, value string) bool {
	switch spec.Type {
	case booleanOption:
		return value == "true" || value == "false"
	case positiveOption:
		number, err := strconv.ParseFloat(value, 64)
		return err == nil && number >= 0
	default:
		return slices.Contains(spec.Values, value)
	}
}

func (o options) flag(id string) bool {
	return o[id] == "true"
}

func (o options) number(id string) float64 {
	number, _ := strconv.ParseFloat(o[id], 64)
	return number
}

func (o options) text(id string) string {
	return o[id]
}

func (o options) indent() string {
	unit := " "
	if o.text(optionIndentType) == "Tab" {
		unit = "\t"
	}
	return strings.Repeat(unit, max(int(o.number(optionIndentCount)), 0))
}
