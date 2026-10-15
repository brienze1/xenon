package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPResponseBody struct {
	Offset int `json:"offset" description:"Character offset to read the body from."`
	Length int `json:"length" description:"Characters to read, at most 65536."`
}

// ReadLength caps the requested length at the most one read may return.
func (m *MCPResponseBody) ReadLength() int {
	return min(m.Length, entity.AIMaxBodyReadBytes)
}

// Slice reads the requested window of the masked body; false means out-of-range arguments.
func (m *MCPResponseBody) Slice(masked string) (entity.AIBodySlice, bool) {
	return entity.NewAIBodySlice(masked, m.Offset, m.ReadLength())
}
