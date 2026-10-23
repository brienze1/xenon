package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type StartRun struct {
	LocalId                   string    `json:"local_id" validate:"required,not_blank"`
	ItemId                    string    `json:"item_id" validate:"required,not_blank"`
	ConfirmedEnvironmentId    string    `json:"confirmed_environment_id"`
	Config                    RunConfig `json:"config"`
	AcknowledgeMissingSecrets bool      `json:"acknowledge_missing_secrets"`
}

func (s *StartRun) ToEntity() entity.RunStart {
	return entity.RunStart{
		LocalID:                   s.LocalId,
		ItemID:                    s.ItemId,
		Config:                    s.Config.ToEntity(),
		ConfirmedEnvironmentID:    s.ConfirmedEnvironmentId,
		AcknowledgeMissingSecrets: s.AcknowledgeMissingSecrets,
	}
}
