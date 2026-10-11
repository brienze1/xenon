package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
)

// Redactable values are logged through LogValue so secrets never reach the log.
type Redactable interface {
	LogValue() any
}

func redact(metadata []any) []any {
	if len(metadata) == 0 {
		return nil
	}
	result := make([]any, 0, len(metadata))
	for _, item := range metadata {
		result = append(result, redactValue(item))
	}
	return result
}

func redactValue(value any) any {
	switch typed := value.(type) {
	case Redactable:
		if isNilPointer(typed) {
			return nil
		}
		return typed.LogValue()
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for key, item := range typed {
			redacted[key] = redactValue(item)
		}
		return redacted
	default:
		return redactSlice(value)
	}
}

func isNilPointer(value any) bool {
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}

func redactSlice(value any) any {
	reflected := reflect.ValueOf(value)
	if reflected.Kind() != reflect.Slice || !holdsRedactable(reflected.Type().Elem()) {
		return value
	}
	redacted := make([]any, 0, reflected.Len())
	for index := range reflected.Len() {
		redacted = append(redacted, redactValue(reflected.Index(index).Interface()))
	}
	return redacted
}

func holdsRedactable(elementType reflect.Type) bool {
	switch elementType.Kind() {
	case reflect.Interface, reflect.Map:
		return true
	case reflect.Slice:
		return elementType.Elem().Kind() != reflect.Uint8
	default:
		return elementType.Implements(reflect.TypeFor[Redactable]())
	}
}

const redactedText = "•••••"

type redactionsKey struct{}

// WithRedactions masks every occurrence of values in the log lines written with the returned context.
func WithRedactions(ctx context.Context, values ...string) context.Context {
	redactions := append(redactionsOf(ctx), nonEmpty(values)...)
	return context.WithValue(ctx, redactionsKey{}, redactions)
}

func redactionsOf(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}
	redactions, _ := ctx.Value(redactionsKey{}).([]string)
	return append([]string{}, redactions...)
}

func nonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func redactLine(ctx context.Context, line []byte) []byte {
	for _, value := range redactionsOf(ctx) {
		line = bytes.ReplaceAll(line, []byte(value), []byte(redactedText))
		encoded, _ := json.Marshal(value)
		line = bytes.ReplaceAll(line, encoded[1:len(encoded)-1], []byte(redactedText))
	}
	return line
}
