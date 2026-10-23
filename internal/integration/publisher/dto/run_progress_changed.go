package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const RunProgressChangedEvent = "runner:progress"

type RunProgressChanged struct {
	LocalId  string      `json:"local_id"`
	RunId    string      `json:"run_id"`
	ItemId   string      `json:"item_id"`
	Type     string      `json:"type"`
	Status   string      `json:"status"`
	Progress RunProgress `json:"progress"`
}

func NewRunProgressChanged(run entity.Run) RunProgressChanged {
	return RunProgressChanged{
		LocalId:  run.Target.LocalID,
		RunId:    run.ID,
		ItemId:   run.Target.ItemID,
		Type:     run.Config.Type.String(),
		Status:   run.Status.String(),
		Progress: NewRunProgress(run.Progress),
	}
}
