package model

import (
	"encoding/json"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ImportProbe struct {
	Info      *ImportProbeInfo `json:"info" yaml:"info"`
	Values    any              `json:"values" yaml:"values"`
	Workspace any              `json:"workspace" yaml:"workspace"`
	ID        any              `json:"id" yaml:"id"`
	Scope     string           `json:"_postman_variable_scope" yaml:"_postman_variable_scope"`
	Kind      string           `json:"$kind" yaml:"$kind"`
	Item      json.RawMessage  `json:"item" yaml:"-"`
	Requests  json.RawMessage  `json:"requests" yaml:"-"`
	Order     json.RawMessage  `json:"order" yaml:"-"`
}

func (i *ImportProbe) ToEntity(name string) entity.ImportProbe {
	probe := i.postmanProbe(name)
	probe.IsJson = true
	probe.HasItems = len(i.Item) > 0
	probe.HasV1Shape = i.Info == nil && (len(i.Requests) > 0 || len(i.Order) > 0)
	return probe
}

func (i *ImportProbe) ToYamlEntity(name string) entity.ImportProbe {
	probe := i.postmanProbe(name)
	probe.IsYaml = true
	return probe
}

func (i *ImportProbe) postmanProbe(name string) entity.ImportProbe {
	probe := entity.ImportProbe{Name: name, Scope: i.Scope, PostmanKind: i.Kind, HasWorkspaceField: i.Workspace != nil}
	if i.Info != nil {
		probe.Schema = i.Info.Schema
	}
	if _, isList := i.Values.([]any); isList {
		probe.HasValues = true
	}
	if id, isText := i.ID.(string); isText {
		probe.DocumentID = id
	}
	return probe
}
