package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/runtype"
)

type RunConfig struct {
	Type          string                 `json:"type" validate:"required,oneof=functional performance"`
	EnvironmentId string                 `json:"environment_id"`
	Rows          []RunRow               `json:"rows" validate:"dive"`
	Functional    RunFunctionalSettings  `json:"functional"`
	Performance   RunPerformanceSettings `json:"performance"`
}

func NewRunConfig(config entity.RunConfig) RunConfig {
	rows := make([]RunRow, 0, len(config.Rows))
	for _, row := range config.Rows {
		rows = append(rows, NewRunRow(row))
	}
	return RunConfig{
		Type:          config.Type.String(),
		EnvironmentId: config.EnvironmentID,
		Rows:          rows,
		Functional:    NewRunFunctionalSettings(config.Functional),
		Performance:   NewRunPerformanceSettings(config.Performance),
	}
}

func (r *RunConfig) ToEntity() entity.RunConfig {
	rows := make([]entity.RunRow, 0, len(r.Rows))
	for _, row := range r.Rows {
		rows = append(rows, row.ToEntity())
	}
	return entity.RunConfig{
		Type:          runtype.GetRunType(r.Type),
		EnvironmentID: r.EnvironmentId,
		Rows:          rows,
		Functional:    r.Functional.ToEntity(),
		Performance:   r.Performance.ToEntity(),
	}
}
