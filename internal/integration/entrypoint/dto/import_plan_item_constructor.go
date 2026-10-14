package dto

func NewImportPlanItem(key, targetLocalId, resolution string, include bool) ImportPlanItem {
	return ImportPlanItem{Key: key, TargetLocalId: targetLocalId, Resolution: resolution, Include: include}
}
