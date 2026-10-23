package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/runtype"
)

type MCPRun struct {
	Functional  *MCPRunFunctional  `json:"functional" description:"Functional settings over the remembered ones."`
	Performance *MCPRunPerformance `json:"performance" description:"Performance settings over the remembered ones."`
	MCPRunTarget
	Type        string      `json:"type" description:"Run type." enum:"functional,performance" validate:"omitempty,oneof=functional performance"`
	Environment string      `json:"environment" description:"Environment name or id, or none."`
	Rows        []MCPRunRow `json:"rows" description:"Requests in run order; default the remembered ones." validate:"dive"`
}

func (m *MCPRun) ToEntity() entity.RunOverrides {
	overrides := entity.RunOverrides{Functional: m.Functional.ToEntity(), Performance: m.Performance.ToEntity()}
	if m.Type != "" {
		runType := runtype.GetRunType(m.Type)
		overrides.Type = &runType
	}
	if m.Environment != "" {
		environment := m.Environment
		overrides.Environment = &environment
	}
	if m.Rows != nil {
		overrides.Rows = make([]entity.RunRow, 0, len(m.Rows))
		for _, row := range m.Rows {
			overrides.Rows = append(overrides.Rows, row.ToEntity())
		}
	}
	return overrides
}
