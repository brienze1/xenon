package entity

type FileChange struct {
	Requests            []RequestChange
	VariableScopes      []ScopeRef
	OwnWrite            bool
	EnvironmentsChanged bool
}

func (f FileChange) IsOwnWrite() bool {
	return f.OwnWrite
}

func (f FileChange) TouchesEnvironments() bool {
	return f.EnvironmentsChanged
}

func (f FileChange) ChangesVariables() bool {
	return f.EnvironmentsChanged || len(f.VariableScopes) > 0
}
