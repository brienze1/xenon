package entity

type RunPlan struct {
	ActiveRun           *Run
	Target              RunTarget
	ActiveEnvironmentID string
	Rows                []RunPlanRow
	Environments        []Environment
	Config              RunConfig
	DataFileMissing     bool
}

func NewRunPlan(target RunTarget, rows []RunPlanRow, config RunConfig, environments []Environment, activeEnvironmentId string) RunPlan {
	return RunPlan{Target: target, Rows: rows, Config: config, Environments: environments, ActiveEnvironmentID: activeEnvironmentId}
}

func (r RunPlan) WithActiveRunOf(runs []Run) RunPlan {
	if active, found := ActiveRunOf(runs, r.Target.ItemID); found {
		r.ActiveRun = &active
	}
	return r
}

func (r RunPlan) WithDataFileMissing(missing bool) RunPlan {
	r.DataFileMissing = missing
	return r
}

// Configured lays the overrides over the remembered configuration; false when the overriding environment is unknown.
func (r RunPlan) Configured(overrides RunOverrides) (RunConfig, bool) {
	config := r.Config.Override(overrides)
	if !overrides.OverridesEnvironment() {
		return config, true
	}
	environmentId, known := r.EnvironmentID(*overrides.Environment)
	config.EnvironmentID = environmentId
	return config, known
}

// EnvironmentID resolves an environment name or id, or NoEnvironmentID; false when no environment matches.
func (r RunPlan) EnvironmentID(ref string) (string, bool) {
	if UsesActiveEnvironment(ref) || UsesNoEnvironment(ref) {
		return ref, true
	}
	for _, environment := range r.Environments {
		if environment.ID == ref || environment.Name == ref {
			return environment.ID, true
		}
	}
	return "", false
}

func (r RunPlan) HasActiveRun() bool {
	return r.ActiveRun != nil
}

func ActiveRunOf(active []Run, itemId string) (Run, bool) {
	for _, run := range active {
		if run.Target.ItemID == itemId {
			return run, true
		}
	}
	return Run{}, false
}
