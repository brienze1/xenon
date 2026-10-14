package dto

type MCPScriptReference struct {
	Members []MCPScriptMember `json:"members"`
}

func NewMCPScriptReference(reference ScriptReference) MCPScriptReference {
	members := make([]MCPScriptMember, 0, len(reference.Members))
	for _, member := range reference.Members {
		members = append(members, NewMCPScriptMember(member))
	}
	return MCPScriptReference{Members: members}
}

func (m MCPScriptReference) LogValue() any {
	return map[string]any{"members": len(m.Members)}
}
