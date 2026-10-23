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
