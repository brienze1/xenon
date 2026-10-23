package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
)

const runVersion = 1

type Run struct {
	StartedAt       time.Time       `json:"started_at"`
	FinishedAt      time.Time       `json:"finished_at"`
	Id              string          `json:"id"`
	Status          string          `json:"status"`
	EnvironmentName string          `json:"environment_name"`
	DataFileName    string          `json:"data_file_name"`
	Target          RunTarget       `json:"target"`
	Config          RunConfig       `json:"config"`
	Rows            []RunRowResult  `json:"rows"`
	Notices         []string        `json:"notices"`
	Performance     *RunPerformance `json:"performance,omitempty"`
	Progress        RunProgress     `json:"progress"`
	Summary         RunSummary      `json:"summary"`
	Version         int             `json:"version"`
}

func NewRun(run entity.Run) Run {
	rows := make([]RunRowResult, 0, len(run.Rows))
	for _, row := range run.Rows {
		rows = append(rows, NewRunRowResult(row))
	}
	return Run{
		Version:         runVersion,
		Id:              run.ID,
		Status:          run.Status.String(),
		StartedAt:       run.StartedAt.UTC(),
		FinishedAt:      run.FinishedAt.UTC(),
		EnvironmentName: run.EnvironmentName,
		DataFileName:    run.DataFileName,
		Target:          NewRunTarget(run.Target),
		Config:          NewRunConfig(run.Config),
		Rows:            rows,
		Notices:         append([]string{}, run.Notices...),
		Performance:     NewRunPerformance(run.Performance),
		Progress:        NewRunProgress(run.Progress),
		Summary:         NewRunSummary(run.Summary),
	}
}

func (r *Run) ToEntity(localId string) entity.Run {
	rows := make([]entity.RunRowResult, 0, len(r.Rows))
	for _, row := range r.Rows {
		rows = append(rows, row.ToEntity())
	}
	return entity.Run{
		ID:              r.Id,
		Status:          runstatus.GetRunStatus(r.Status),
		StartedAt:       r.StartedAt,
		FinishedAt:      r.FinishedAt,
		EnvironmentName: r.EnvironmentName,
		DataFileName:    r.DataFileName,
		Target:          r.Target.ToEntity(localId),
		Config:          r.Config.ToEntity(),
		Rows:            rows,
		Notices:         append([]string{}, r.Notices...),
		Performance:     r.Performance.ToEntity(),
		Progress:        r.Progress.ToEntity(),
		Summary:         r.Summary.ToEntity(),
	}
}
