package dto

import (
	"encoding/json"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type GraphQlBody struct {
	Query     string          `json:"query"`
	Variables json.RawMessage `json:"variables"`
}

func NewGraphQlBody(body entity.BodyGraphQl) GraphQlBody {
	variables := json.RawMessage("{}")
	if trimmed := strings.TrimSpace(body.Variables); json.Valid([]byte(trimmed)) {
		variables = json.RawMessage(trimmed)
	}
	return GraphQlBody{Query: body.Query, Variables: variables}
}
