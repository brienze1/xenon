package dto

import (
	"encoding/json"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type HeaderList struct {
	Rows        []KeyValue
	IsRawString bool
}

func (h *HeaderList) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return json.Unmarshal(data, &h.Rows)
	}
	h.IsRawString = true
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		key, value, found := strings.Cut(line, ":")
		if found && strings.TrimSpace(key) != "" {
			h.Rows = append(h.Rows, KeyValue{Key: strings.TrimSpace(key), Value: strings.TrimSpace(value)})
		}
	}
	return nil
}

func (h *HeaderList) ToEntity() []entity.KeyValue {
	return keyValuesToEntity(h.Rows)
}

func (h *HeaderList) Passthrough(fields map[string]any) map[string]any {
	if h.IsRawString {
		fields["header"] = map[string]any{"raw": true}
		return fields
	}
	return withList(fields, "header", keyValuePassthrough(h.Rows))
}
