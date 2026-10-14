package dto

func NewSyncSessionReference(sessionId string) SyncSessionReference {
	return SyncSessionReference{SessionId: sessionId}
}
