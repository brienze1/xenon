package dto

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
)

func decodeRaw(data []byte) map[string]any {
	var raw map[string]any
	_ = json.Unmarshal(data, &raw)
	return raw
}

func without(raw map[string]any, keys ...string) map[string]any {
	rest := map[string]any{}
	for key, value := range raw {
		if !slices.Contains(keys, key) {
			rest[key] = value
		}
	}
	return rest
}

func withEntry(fields map[string]any, key string, value map[string]any) map[string]any {
	if len(value) > 0 {
		fields[key] = value
	}
	return fields
}

func withList(fields map[string]any, key string, value []any) map[string]any {
	for _, item := range value {
		if entry, isMap := item.(map[string]any); !isMap || len(entry) > 0 {
			fields[key] = value
			return fields
		}
	}
	return fields
}

func text(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		encoded, _ := json.Marshal(typed)
		return string(encoded)
	}
}

func lines(value any) string {
	if list, isList := value.([]any); isList {
		parts := make([]string, 0, len(list))
		for _, line := range list {
			parts = append(parts, text(line))
		}
		return strings.Join(parts, "\n")
	}
	return text(value)
}

func asMap(value any) map[string]any {
	if fields, isMap := value.(map[string]any); isMap {
		return fields
	}
	return map[string]any{}
}

func asList(value any) []any {
	if list, isList := value.([]any); isList {
		return list
	}
	return []any{}
}

func mapAt(list []any, index int) map[string]any {
	if index < len(list) {
		return asMap(list[index])
	}
	return map[string]any{}
}

func sortedKeys(fields map[string]any) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func descriptionValue(content string, passthrough any) any {
	if fields := asMap(passthrough); len(fields) > 0 {
		return newObject().set("content", content).merged(fields)
	}
	if content == "" {
		return nil
	}
	return content
}

func itemId(passthrough map[string]any, id string, xenonIds bool) any {
	if value, found := passthrough["id"]; found {
		return value
	}
	if xenonIds {
		return id
	}
	return nil
}

func scriptLines(text string) []string {
	return strings.Split(text, "\n")
}
