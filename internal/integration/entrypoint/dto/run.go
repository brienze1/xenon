package dto

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const runTimeLayout = "2006-01-02T15:04:05.000Z"

type Run struct {
	Performance     any            `json:"performance"`
	Id              string         `json:"id"`
	EnvironmentName string         `json:"environment_name"`
	DataFileName    string         `json:"data_file_name"`
	Status          string         `json:"status"`
	StartedAt       string         `json:"started_at"`
	FinishedAt      string         `json:"finished_at"`
	Target          RunTarget      `json:"target"`
	Config          RunConfig      `json:"config"`
	Rows            []RunRowResult `json:"rows"`
	Notices         []string       `json:"notices"`
	Progress        RunProgress    `json:"progress"`
	Summary         RunSummary     `json:"summary"`
}

func NewRun(run entity.Run) Run {
	rows := make([]RunRowResult, 0, len(run.Rows))
	for _, row := range run.Rows {
		rows = append(rows, NewRunRowResult(row))
	}
	return Run{
		Id:              run.ID,
		Target:          NewRunTarget(run.Target),
		Config:          NewRunConfig(run.Config),
		EnvironmentName: run.EnvironmentName,
		DataFileName:    run.DataFileName,
		Status:          run.Status.String(),
		StartedAt:       runTime(run.StartedAt),
		FinishedAt:      runTime(run.FinishedAt),
		Progress:        NewRunProgress(run.Progress),
		Summary:         NewRunSummary(run.Summary),
		Rows:            rows,
		Notices:         append([]string{}, run.Notices...),
	}
}

func runTime(at time.Time) string {
	if at.IsZero() {
		return ""
	}
	return at.UTC().Format(runTimeLayout)
}
