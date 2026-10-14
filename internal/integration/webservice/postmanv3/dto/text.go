package dto

import (
	"fmt"
	"maps"
	"strconv"
)

func text(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprint(typed)
	}
}

func copied(extra map[string]any) map[string]any {
	result := make(map[string]any, len(extra)+4)
	maps.Copy(result, extra)
	return result
}

func setIfPresent(target map[string]any, key string, value any) {
	switch typed := value.(type) {
	case nil:
	case string:
		if typed != "" {
			target[key] = typed
		}
	case []any:
		if len(typed) > 0 {
			target[key] = typed
		}
	case map[string]any:
		if len(typed) > 0 {
			target[key] = typed
		}
	default:
		target[key] = typed
	}
}
