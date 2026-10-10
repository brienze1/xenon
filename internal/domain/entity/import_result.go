package entity

type ImportResult struct {
	Items      []ImportItemResult
	Workspaces []ImportWorkspaceResult
	Warnings   []ImportWarning
	PmUsages   []PmUsage
	Failed     []ImportFailure
}

func NewImportResult() ImportResult {
	return ImportResult{Items: []ImportItemResult{}, Workspaces: []ImportWorkspaceResult{}, Warnings: []ImportWarning{}, PmUsages: []PmUsage{}, Failed: []ImportFailure{}}
}

func (i ImportResult) WithOutcome(outcome ImportOutcome) ImportResult {
	i.Items = append(i.Items, outcome.Item)
	i.Workspaces = i.workspacesWith(outcome)
	if outcome.Failure != nil {
		i.Failed = append(i.Failed, *outcome.Failure)
		return i
	}
	for _, warning := range outcome.Warnings {
		i.Warnings = append(i.Warnings, warning.InWorkspace(outcome.Item.LocalID))
	}
	usages := append([]PmUsage{}, i.PmUsages...)
	for _, usage := range outcome.PmUsages {
		usages = append(usages, usage.InWorkspace(outcome.Item.LocalID))
	}
	i.PmUsages = MergePmUsages(usages)
	return i
}

func (i ImportResult) WithSourceWarnings(warnings []ImportWarning) ImportResult {
	i.Warnings = append(append([]ImportWarning{}, i.Warnings...), warnings...)
	return i
}

func (i ImportResult) workspacesWith(outcome ImportOutcome) []ImportWorkspaceResult {
	workspaces := append([]ImportWorkspaceResult{}, i.Workspaces...)
	index := len(workspaces)
	for position, workspace := range workspaces {
		if workspace.LocalID == outcome.Item.LocalID {
			index = position
		}
	}
	if index == len(workspaces) {
		workspaces = append(workspaces, ImportWorkspaceResult{LocalID: outcome.Item.LocalID})
	}
	workspaces[index] = workspaces[index].with(outcome)
	return workspaces
}

func (i ImportResult) LocalIDs() []string {
	ids := make([]string, 0, len(i.Workspaces))
	for _, workspace := range i.Workspaces {
		ids = append(ids, workspace.LocalID)
	}
	return ids
}

func (i ImportResult) LogValue() any {
	return map[string]any{"items": len(i.Items), "warnings": len(i.Warnings), "failed": len(i.Failed)}
}
