package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunSummaryView struct {
	Id              string      `json:"id"`
	ItemId          string      `json:"item_id"`
	TargetKind      string      `json:"target_kind"`
	TargetName      string      `json:"target_name"`
	Type            string      `json:"type"`
	Status          string      `json:"status"`
	EnvironmentName string      `json:"environment_name"`
	StartedAt       string      `json:"started_at"`
	FinishedAt      string      `json:"finished_at"`
	Progress        RunProgress `json:"progress"`
	Summary         RunSummary  `json:"summary"`
}

func NewRunSummaryView(run entity.Run) RunSummaryView {
	return RunSummaryView{
		Id:              run.ID,
		ItemId:          run.Target.ItemID,
		TargetKind:      run.Target.Kind.String(),
		TargetName:      run.Target.Name,
		Type:            run.Config.Type.String(),
		Status:          run.Status.String(),
		EnvironmentName: run.EnvironmentName,
		StartedAt:       runTime(run.StartedAt),
		FinishedAt:      runTime(run.FinishedAt),
		Progress:        NewRunProgress(run.Progress),
		Summary:         NewRunSummary(run.Summary),
	}
}

func NewRunSummaryViews(runs []entity.Run) []RunSummaryView {
	views := make([]RunSummaryView, 0, len(runs))
	for _, run := range runs {
		views = append(views, NewRunSummaryView(run))
	}
	return views
}
