package entity

const globalsImportPath = "Globals"

type ImportVariablesWrite struct {
	Environment   *Environment
	Replaced      *Environment
	Globals       *Workspace
	Secrets       map[string]string
	KeptVariables []VariableRow
	Added         int
}

func (i ImportVariablesWrite) IsEmpty() bool {
	return i.Environment == nil && i.Globals == nil
}

func (i ImportVariablesWrite) Scope() ScopeRef {
	if i.Environment != nil {
		return i.Environment.ScopeRef()
	}
	return NewGlobalsScopeRef()
}

func (i ImportVariablesWrite) Counts() ImportCounts {
	if i.Environment != nil {
		return ImportCounts{Environments: 1, Variables: len(i.Environment.Variables)}
	}
	return ImportCounts{Variables: i.Added}
}

func (i ImportVariablesWrite) target(name string) (string, string, string) {
	if i.Environment != nil {
		return i.Environment.ID, i.Environment.Name, i.Environment.Name
	}
	return i.Globals.ID, name, globalsImportPath
}
