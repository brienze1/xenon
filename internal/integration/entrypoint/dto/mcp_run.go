package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
)

type MCPRun struct {
	Functional  *MCPRunFunctional  `json:"functional" description:"Functional settings over the remembered ones."`
	Performance *MCPRunPerformance `json:"performance" description:"Performance settings over the remembered ones."`
	MCPRunTarget
	Type        string      `json:"type" description:"Run type." enum:"functional,performance"`
	Environment string      `json:"environment" description:"Environment name or id, or none."`
	Rows        []MCPRunRow `json:"rows" description:"Requests in run order; default the remembered ones."`
}

// RunConfig lays the arguments over the remembered configuration; false when the environment is unknown.
func (m *MCPRun) RunConfig(plan RunPlan) (RunConfig, bool) {
	config := plan.Config
	if m.Rows != nil {
		config.Rows = make([]RunRow, 0, len(m.Rows))
		for _, row := range m.Rows {
			config.Rows = append(config.Rows, row.RunRow())
		}
	}
	if m.Type != "" {
		config.Type = m.Type
	}
	config.Functional = m.Functional.Over(config.Functional)
	config.Performance = m.Performance.Over(config.Performance)
	environmentId, known := m.environmentId(plan.Environments)
	if m.Environment != "" {
		config.EnvironmentId = environmentId
	}
	return config, known
}

func (m *MCPRun) StartRun(localId, itemId, confirmedEnvironmentId string, config RunConfig) StartRun {
	return StartRun{LocalId: localId, ItemId: itemId, ConfirmedEnvironmentId: confirmedEnvironmentId, Config: config}
}

func (m *MCPRun) Changes(config RunConfig) []enums.AIChangeKind {
	settings := config.ToEntity()
	changes := make([]enums.AIChangeKind, 0, 2)
	if settings.SavesCookies() {
		changes = append(changes, aichangekind.Cookies)
	}
	if !settings.IsPerformance() && settings.Functional.KeepVariableValues {
		changes = append(changes, aichangekind.Variables)
	}
	return changes
}

func (m *MCPRun) environmentId(environments []EnvironmentSummary) (string, bool) {
	if m.Environment == "" || m.Environment == entity.NoEnvironmentID {
		return m.Environment, true
	}
	for _, environment := range environments {
		if environment.Id == m.Environment || environment.Name == m.Environment {
			return environment.Id, true
		}
	}
	return "", false
}
