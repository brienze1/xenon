package steps

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var placeholder = regexp.MustCompile(`\{\{([^}]+)\}\}`)

func (t *testUtils) substitute(text string) string {
	return placeholder.ReplaceAllStringFunc(text, func(match string) string {
		name := placeholder.FindStringSubmatch(match)[1]
		if value, found := t.vars[name]; found {
			return value
		}
		return match
	})
}

func (t *testUtils) substituteJson(text string) string {
	return placeholder.ReplaceAllStringFunc(text, func(match string) string {
		name := placeholder.FindStringSubmatch(match)[1]
		value, found := t.vars[name]
		if !found {
			return match
		}
		encoded, _ := json.Marshal(value)
		return string(encoded[1 : len(encoded)-1])
	})
}

func (t *testUtils) workspacePath(relative string) string {
	return filepath.Join(t.workspaceRoot, filepath.FromSlash(t.substitute(relative)))
}

func lookup(data any, path string) (any, bool) {
	current := data
	for _, segment := range strings.Split(path, ".") {
		switch node := current.(type) {
		case map[string]any:
			value, found := node[segment]
			if !found {
				return nil, false
			}
			current = value
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(node) {
				return nil, false
			}
			current = node[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func stringify(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case map[string]any, []any:
		encoded, _ := json.Marshal(typed)
		return string(encoded)
	default:
		return fmt.Sprint(typed)
	}
}

func decodeJson(raw []byte) (any, error) {
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("invalid JSON %q: %w", raw, err)
	}
	return data, nil
}

func fieldOf(raw []byte, path string) (any, error) {
	data, err := decodeJson(raw)
	if err != nil {
		return nil, err
	}
	value, found := lookup(data, path)
	if !found {
		return nil, fmt.Errorf("field %q not found in %s", path, raw)
	}
	return value, nil
}

func expectEqual(actual, expected, what string) error {
	if actual != expected {
		return fmt.Errorf("%s: expected %q, got %q", what, expected, actual)
	}
	return nil
}
