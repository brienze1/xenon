package entity

type ImportWorkspaceResult struct {
	LocalID  string
	Created  ImportCounts
	Replaced int
	Skipped  int
}

func (i ImportWorkspaceResult) with(outcome ImportOutcome) ImportWorkspaceResult {
	i.Created = i.Created.Add(outcome.Created)
	if outcome.Replaced {
		i.Replaced++
	}
	if outcome.Skipped {
		i.Skipped++
	}
	return i
}
