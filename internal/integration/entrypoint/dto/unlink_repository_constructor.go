package dto

func NewUnlinkRepository(localId, targetFolder string) UnlinkRepository {
	return UnlinkRepository{LocalId: localId, TargetFolder: targetFolder}
}
