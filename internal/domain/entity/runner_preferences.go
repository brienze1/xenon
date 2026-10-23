package entity

type RunnerPreferences struct {
	Config RunConfig
}

func NewRunnerPreferences(config RunConfig) RunnerPreferences {
	return RunnerPreferences{Config: config}
}

func (r RunnerPreferences) RememberedConfig() RunConfig {
	return r.Config.Completed()
}
