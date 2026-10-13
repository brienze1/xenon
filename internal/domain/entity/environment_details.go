package entity

type EnvironmentDetails struct {
	Table       VariableTable
	Environment Environment
}

func NewEnvironmentDetails(environment Environment, local LocalVariables, vault Vault) EnvironmentDetails {
	return EnvironmentDetails{Environment: environment, Table: NewVariableTable(environment.ScopeRef(), environment.Variables, local, vault)}
}

func (e EnvironmentDetails) LogValue() any {
	return map[string]any{"environment": e.Environment.LogValue(), "table": e.Table.LogValue()}
}
