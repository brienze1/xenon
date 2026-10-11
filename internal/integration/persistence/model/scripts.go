package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Scripts struct {
	PreRequest   string   `yaml:"preRequest"`
	PostResponse string   `yaml:"postResponse"`
	Disabled     []string `yaml:"disabled"`
}

func NewScripts(scripts entity.Scripts) Scripts {
	model := Scripts{PreRequest: scripts.PreRequest, PostResponse: scripts.PostResponse, Disabled: make([]string, 0, len(scripts.Disabled))}
	for _, phase := range scripts.Disabled {
		model.Disabled = append(model.Disabled, phase.String())
	}
	return model
}

func (s *Scripts) ToEntity() entity.Scripts {
	scripts := entity.Scripts{PreRequest: s.PreRequest, PostResponse: s.PostResponse}
	for _, raw := range s.Disabled {
		if phase := scriptphase.GetScriptPhase(raw); phase != scriptphase.Unknown {
			scripts.Disabled = append(scripts.Disabled, phase)
		}
	}
	return scripts
}

func (s *Scripts) Node() *yaml.Node {
	fields := []yamlnode.Field{yamlnode.String("preRequest", s.PreRequest), yamlnode.String("postResponse", s.PostResponse)}
	if len(s.Disabled) > 0 {
		fields = append(fields, yamlnode.StringList("disabled", s.Disabled))
	}
	return yamlnode.MappingNode(fields...)
}

func scriptsField(scripts Scripts, edited bool) yamlnode.Field {
	return scriptsReserved(yamlnode.Reserved(scriptsKey, scripts.Node()), edited)
}

func scriptsReserved(field yamlnode.Field, edited bool) yamlnode.Field {
	if edited {
		return overwritten(field)
	}
	return field
}
