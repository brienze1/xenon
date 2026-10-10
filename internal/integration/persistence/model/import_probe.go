package model

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ImportProbe struct {
	Info *struct {
		Schema string `json:"schema"`
	} `json:"info"`
	Scope    string          `json:"_postman_variable_scope"`
	Item     json.RawMessage `json:"item"`
	Requests json.RawMessage `json:"requests"`
	Order    json.RawMessage `json:"order"`
}

func NewImportProbe(name string, data []byte, isJsonFile bool) entity.ImportProbe {
	probe := entity.ImportProbe{Name: name}
	trimmed := bytes.TrimSpace(data)
	if !isJsonFile && !bytes.HasPrefix(trimmed, []byte("{")) && !bytes.HasPrefix(trimmed, []byte("[")) {
		return probe
	}
	var document ImportProbe
	if err := json.Unmarshal(trimmed, &document); err != nil {
		var syntax *json.SyntaxError
		probe.IsMalformed = errors.As(err, &syntax)
		return probe
	}
	return document.ToEntity(probe)
}

func (i *ImportProbe) ToEntity(probe entity.ImportProbe) entity.ImportProbe {
	probe.IsJson = true
	probe.Scope = i.Scope
	probe.HasItems = len(i.Item) > 0
	if i.Info != nil {
		probe.Schema = i.Info.Schema
	}
	probe.HasV1Shape = i.Info == nil && (len(i.Requests) > 0 || len(i.Order) > 0)
	return probe
}
