package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type OAuth2Param struct {
	Enabled *bool  `yaml:"enabled" json:"enabled,omitempty"`
	Key     string `yaml:"key" json:"key"`
	Value   string `yaml:"value" json:"value"`
	SendAs  string `yaml:"send_as" json:"send_as"`
}

func newOAuth2Params(params []entity.OAuth2Param) []OAuth2Param {
	models := make([]OAuth2Param, 0, len(params))
	for _, param := range params {
		models = append(models, OAuth2Param{Key: param.Key, Value: param.Value, Enabled: &param.Enabled, SendAs: param.SendAs})
	}
	return models
}

func oauth2ParamsToEntity(params []OAuth2Param, defaultSendAs string) []entity.OAuth2Param {
	entities := make([]entity.OAuth2Param, 0, len(params))
	for _, param := range params {
		sendAs := param.SendAs
		if sendAs == "" {
			sendAs = defaultSendAs
		}
		entities = append(entities, entity.OAuth2Param{Key: param.Key, Value: param.Value, SendAs: sendAs, Enabled: param.Enabled == nil || *param.Enabled})
	}
	return entities
}

func oauth2ParamNodes(params []entity.OAuth2Param) []*yaml.Node {
	nodes := make([]*yaml.Node, 0, len(params))
	for _, param := range params {
		nodes = append(nodes, yamlnode.MappingNode(
			yamlnode.String("key", param.Key),
			yamlnode.String("value", param.Value),
			yamlnode.Bool("enabled", param.Enabled),
			yamlnode.String("send_as", param.SendAs),
		))
	}
	return nodes
}
