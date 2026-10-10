package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

func newVariableRows(rows []entity.VariableRow) []VariableRow {
	models := make([]VariableRow, 0, len(rows))
	for _, row := range rows {
		models = append(models, NewVariableRow(row))
	}
	return models
}

func variableRowsToEntity(models []VariableRow) []entity.VariableRow {
	rows := make([]entity.VariableRow, 0, len(models))
	for index := range models {
		rows = append(rows, models[index].ToEntity())
	}
	return rows
}

func variablesField(models []VariableRow) yamlnode.Field {
	nodes := make([]*yaml.Node, 0, len(models))
	for index := range models {
		nodes = append(nodes, models[index].Node())
	}
	return yamlnode.List("variables", nodes)
}
