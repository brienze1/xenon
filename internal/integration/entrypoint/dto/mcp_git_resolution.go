package dto

type MCPGitResolution struct {
	Id     string `json:"id" description:"Conflicting item id." validated_as:"item_id"`
	Choice string `json:"choice" description:"How to resolve the conflict." enum:"mine,theirs,both,keep,delete" validated_as:"resolution"`
	Field  string `json:"field" description:"Field path inside the item, for a field-level conflict."`
}

func (m *MCPGitResolution) ResolveSyncConflict(sessionId string) ResolveSyncConflict {
	return NewResolveSyncConflict(sessionId, m.Id, m.Choice, m.Field)
}
