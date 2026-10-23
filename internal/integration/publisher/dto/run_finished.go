package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const RunFinishedEvent = "runner:finished"

type RunFinished struct {
	LocalId    string     `json:"local_id"`
	RunId      string     `json:"run_id"`
	ItemId     string     `json:"item_id"`
	TargetName string     `json:"target_name"`
	Type       string     `json:"type"`
	Status     string     `json:"status"`
	Summary    RunSummary `json:"summary"`
}

func NewRunFinished(run entity.Run) RunFinished {
	return RunFinished{
		LocalId:    run.Target.LocalID,
		RunId:      run.ID,
		ItemId:     run.Target.ItemID,
		TargetName: run.Target.Name,
		Type:       run.Config.Type.String(),
		Status:     run.Status.String(),
		Summary:    NewRunSummary(run.Summary),
	}
}
