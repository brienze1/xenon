package mcp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	PageSize         = 100
	InlineBodyBytes  = 16384
	MaxBodyReadBytes = 65536
	cursorPrefix     = "o:"
	itemsField       = "items"
	mcpTabPrefix     = "mcp:"
	emptyArguments   = "{}"
)

type List[T any] struct {
	Next  string `json:"next,omitempty"`
	Items []T    `json:"items"`
}

type BodySlice struct {
	Text      string `json:"text"`
	Offset    int    `json:"offset"`
	Size      int    `json:"size"`
	Truncated bool   `json:"truncated"`
}

// Paginate returns up to 100 items after the opaque cursor and the cursor of the next page.
func Paginate[T any](tool string, items []T, cursor string) (List[T], error) {
	offset, err := cursorOffset(tool, cursor)
	if err != nil {
		return List[T]{}, err
	}
	start := min(offset, len(items))
	end := min(start+PageSize, len(items))
	page := List[T]{Items: append(make([]T, 0, end-start), items[start:end]...)}
	if end < len(items) {
		page.Next = encodeCursor(end)
	}
	return page, nil
}

// InlineBody caps an already masked body at 16 KB on a rune boundary and reports its full masked size.
func InlineBody(masked string) (string, bool, int) {
	if len(masked) <= InlineBodyBytes {
		return masked, false, len(masked)
	}
	return masked[:runeBoundary(masked, InlineBodyBytes)], true, len(masked)
}

// SliceBody reads length bytes (16 KB by default, 64 KB at most) of an already masked body from offset.
func SliceBody(tool, masked string, offset, length int) (BodySlice, error) {
	if offset < 0 || length < 0 || length > MaxBodyReadBytes {
		return BodySlice{}, invalidArguments(tool, "offset must be 0 or more and length at most "+strconv.Itoa(MaxBodyReadBytes))
	}
	if length == 0 {
		length = InlineBodyBytes
	}
	start := runeBoundary(masked, min(offset, len(masked)))
	end := runeBoundary(masked, min(start+length, len(masked)))
	return BodySlice{Text: masked[start:end], Offset: start, Size: len(masked), Truncated: end < len(masked)}, nil
}

func encodeJSON(payload any) (string, error) {
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
			if pruned := omitEmpty(field, key == itemsField); pruned != nil {
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

func runeBoundary(text string, index int) int {
	for index > 0 && index < len(text) && !utf8.RuneStart(text[index]) {
		index--
	}
	return index
}

func encodeCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(cursorPrefix + strconv.Itoa(offset)))
}

func cursorOffset(tool, cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil || !strings.HasPrefix(string(decoded), cursorPrefix) {
		return 0, invalidArguments(tool, "unknown cursor")
	}
	offset, err := strconv.Atoi(strings.TrimPrefix(string(decoded), cursorPrefix))
	if err != nil || offset < 0 {
		return 0, invalidArguments(tool, "unknown cursor")
	}
	return offset, nil
}
