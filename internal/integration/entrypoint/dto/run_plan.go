package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunPlan struct {
	ActiveRun           *RunSummaryView      `json:"active_run"`
	ActiveEnvironmentId string               `json:"active_environment_id"`
	Target              RunTarget            `json:"target"`
	Rows                []RunPlanRow         `json:"rows"`
	Environments        []EnvironmentSummary `json:"environments"`
	Config              RunConfig            `json:"config"`
	DataFileMissing     bool                 `json:"data_file_missing"`
}

func NewRunPlan(plan entity.RunPlan) RunPlan {
	rows := make([]RunPlanRow, 0, len(plan.Rows))
	for _, row := range plan.Rows {
		rows = append(rows, NewRunPlanRow(row))
	}
	environments := make([]EnvironmentSummary, 0, len(plan.Environments))
	for _, environment := range plan.Environments {
		environments = append(environments, NewEnvironmentSummary(environment))
	}
	result := RunPlan{
		Target:              NewRunTarget(plan.Target),
		Rows:                rows,
		Config:              NewRunConfig(plan.Config),
		ActiveEnvironmentId: plan.ActiveEnvironmentID,
		Environments:        environments,
		DataFileMissing:     plan.DataFileMissing,
	}
	if plan.ActiveRun != nil {
		active := NewRunSummaryView(*plan.ActiveRun)
		result.ActiveRun = &active
	}
	return result
}
