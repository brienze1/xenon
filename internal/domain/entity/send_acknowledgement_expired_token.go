package entity

func (s SendAcknowledgement) RejectsExpiredToken(expired bool) bool {
	return expired && !s.AcceptsExpiredToken()
}
