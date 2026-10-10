package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

var placeholder = regexp.MustCompile(`\{\{([^}]+)\}\}`)

var byteOrderMark = []byte{0xEF, 0xBB, 0xBF}

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

func (t *testUtils) substituteFor(path, text string) string {
	if strings.EqualFold(filepath.Ext(path), ".json") {
		return t.substituteJson(text)
	}
	return t.substitute(text)
}

func (t *testUtils) workspacePath(relative string) string {
	return filepath.Join(t.workspaceRoot, filepath.FromSlash(t.substitute(relative)))
}

func (t *testUtils) appDataPath(relative string) string {
	return filepath.Join(t.dataDir, filepath.FromSlash(t.substitute(relative)))
}

func lookup(data any, path string) (any, bool) {
	current := data
	if path == "" {
		return current, true
	}
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

const generatedAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz\n"

func generatedBody(size int) []byte {
	body := make([]byte, size)
	for index := range body {
		body[index] = generatedAlphabet[index%len(generatedAlphabet)]
	}
	return body
}

func resourcePath(name string) string {
	return filepath.Join("..", "resources", filepath.FromSlash(name))
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

func resourceFileField(name, path string) (any, error) {
	content, err := os.ReadFile(resourcePath(name))
	if err != nil {
		return nil, err
	}
	var data any
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("resource %s is not YAML or JSON: %w", name, err)
	}
	value, found := lookup(data, path)
	if !found {
		return nil, fmt.Errorf("field %q not found in the resource %s", path, name)
	}
	return value, nil
}

func matchFields(actual, expected any, path string) error {
	switch want := expected.(type) {
	case map[string]any:
		for key, value := range want {
			child, found := lookup(actual, key)
			if !found {
				return fmt.Errorf("field %q is missing", joinFieldPath(path, key))
			}
			if err := matchFields(child, value, joinFieldPath(path, key)); err != nil {
				return err
			}
		}
		return nil
	case []any:
		got, isList := actual.([]any)
		if !isList || len(got) != len(want) {
			return fmt.Errorf("field %q: expected %d items, got %s", path, len(want), stringify(actual))
		}
		for index := range want {
			if err := matchFields(got[index], want[index], joinFieldPath(path, strconv.Itoa(index))); err != nil {
				return err
			}
		}
		return nil
	default:
		return expectEqual(stringify(actual), stringify(want), "field "+strconv.Quote(path))
	}
}

func joinFieldPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}
