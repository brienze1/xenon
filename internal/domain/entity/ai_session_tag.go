package entity

const aiSessionTagLength = 6

func AISessionTag(sessionId string) string {
	if len(sessionId) <= aiSessionTagLength {
		return sessionId
	}
	return sessionId[:aiSessionTagLength]
}
