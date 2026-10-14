package dto

type MCPScriptMember struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

func NewMCPScriptMember(member ScriptMember) MCPScriptMember {
	return MCPScriptMember{Path: member.Path, Kind: member.Kind}
}
