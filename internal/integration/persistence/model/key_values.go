package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func newKeyValues(rows []entity.KeyValue) []KeyValue {
	models := make([]KeyValue, 0, len(rows))
	for _, row := range rows {
		models = append(models, NewKeyValue(row))
	}
	return models
}

func keyValuesToEntity(models []KeyValue) []entity.KeyValue {
	rows := make([]entity.KeyValue, 0, len(models))
	for index := range models {
		rows = append(rows, models[index].ToEntity())
	}
	return rows
}

func keyValueNodes(models []KeyValue) []*yaml.Node {
	nodes := make([]*yaml.Node, 0, len(models))
	for index := range models {
		nodes = append(nodes, models[index].Node())
	}
	return nodes
}
