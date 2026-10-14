package dto

import (
	"bytes"
	"encoding/json"
	"strings"
)

const (
	mcpItemsField  = "items"
	mcpErrorsField = "errors"
)

// EncodeMCPText encodes payload as compact JSON without empty strings, maps or arrays, except under items and errors.
func EncodeMCPText(payload any) (string, error) {
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(payload); err != nil {
		return "", err
	}
	var decoded any
	decoder := json.NewDecoder(&encoded)
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return "", err
	}
	var compacted bytes.Buffer
	encoder = json.NewEncoder(&compacted)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(omitEmpty(decoded, true)); err != nil {
		return "", err
	}
	return strings.TrimSuffix(compacted.String(), "\n"), nil
}

func omitEmpty(node any, keep bool) any {
	switch value := node.(type) {
	case map[string]any:
		compacted := make(map[string]any, len(value))
		for key, field := range value {
			if pruned := omitEmpty(field, key == mcpItemsField || key == mcpErrorsField); pruned != nil {
				compacted[key] = pruned
			}
		}
		if len(compacted) == 0 && !keep {
			return nil
		}
		return compacted
	case []any:
		if len(value) == 0 && !keep {
			return nil
		}
		items := make([]any, 0, len(value))
		for _, item := range value {
			items = append(items, omitEmptyItem(item))
		}
		return items
	case string:
		if value == "" && !keep {
			return nil
		}
		return value
	default:
		return value
	}
}

func omitEmptyItem(item any) any {
	if compacted := omitEmpty(item, true); compacted != nil {
		return compacted
	}
	return item
}
