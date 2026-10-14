package dto

type MCPEnvInspection struct {
	Refs []MCPEnvReference `json:"refs"`
}

func NewMCPEnvInspection(inspection VariableInspection) MCPEnvInspection {
	refs := make([]MCPEnvReference, 0, len(inspection.Variables))
	for _, usage := range inspection.Variables {
		refs = append(refs, NewMCPEnvReference(usage))
	}
	return MCPEnvInspection{Refs: refs}
}
