package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SaveRunnerPreferences struct {
	LocalId string    `json:"local_id" validate:"required,not_blank"`
	ItemId  string    `json:"item_id" validate:"required,not_blank"`
	Config  RunConfig `json:"config"`
}

func (s *SaveRunnerPreferences) ToEntity() entity.RunnerPreferences {
	return entity.NewRunnerPreferences(s.Config.ToEntity())
}
