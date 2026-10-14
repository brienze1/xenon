package dto

func NewResolveSyncConflict(sessionId, itemId, resolution, fieldPath string) ResolveSyncConflict {
	return ResolveSyncConflict{SessionId: sessionId, ItemId: itemId, Resolution: resolution, FieldPath: fieldPath}
}
