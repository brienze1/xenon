package entity

// IncludedInAIImport leaves whole workspaces out of an import an AI client runs.
func (i ImportItem) IncludedInAIImport() bool {
	return !i.IsWorkspace()
}
