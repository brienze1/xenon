package model

import "github.com/brienze1/xenon/internal/domain/entity"

type RunnerPreferences map[string]RunConfig

func NewRunnerPreferences(preferences map[string]entity.RunnerPreferences) RunnerPreferences {
	runners := make(RunnerPreferences, len(preferences))
	for itemId, remembered := range preferences {
		runners[itemId] = NewRunConfig(remembered.Config)
	}
	return runners
}

func (r RunnerPreferences) ToEntity() map[string]entity.RunnerPreferences {
	preferences := make(map[string]entity.RunnerPreferences, len(r))
	for itemId, config := range r {
		preferences[itemId] = entity.NewRunnerPreferences(config.ToEntity())
	}
	return preferences
}
