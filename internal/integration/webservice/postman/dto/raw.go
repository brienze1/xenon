package dto

import (
	"encoding/json"
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
		if !contains(keys, key) {
			rest[key] = value
		}
	}
	return rest
}

func contains(keys []string, key string) bool {
	for _, candidate := range keys {
		if candidate == key {
			return true
		}
	}
	return false
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
