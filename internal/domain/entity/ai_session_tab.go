package entity

const aiSessionTabPrefix = "mcp:"

func AISessionTabID(sessionId string) string {
	return aiSessionTabPrefix + sessionId
}
