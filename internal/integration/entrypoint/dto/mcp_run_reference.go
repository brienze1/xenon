package dto

type MCPRunReference struct {
	RunId string `json:"run_id" description:"Run id." required_for:"stop,get" validated_as:"run_id"`
}

func (m *MCPRunReference) HasRunId() bool {
	return m.RunId != ""
}

func (m *MCPRunReference) RunReference(localId string) RunReference {
	return RunReference{LocalId: localId, RunId: m.RunId}
}
