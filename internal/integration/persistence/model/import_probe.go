package model

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ImportProbe struct {
	Info     *ImportProbeInfo `json:"info"`
	Scope    string           `json:"_postman_variable_scope"`
	Item     json.RawMessage  `json:"item"`
	Requests json.RawMessage  `json:"requests"`
	Order    json.RawMessage  `json:"order"`
}

func (i *ImportProbe) ToEntity(name string) entity.ImportProbe {
	probe := entity.ImportProbe{Name: name, IsJson: true, Scope: i.Scope, HasItems: len(i.Item) > 0}
	if i.Info != nil {
		probe.Schema = i.Info.Schema
	}
	probe.HasV1Shape = i.Info == nil && (len(i.Requests) > 0 || len(i.Order) > 0)
	return probe
}
