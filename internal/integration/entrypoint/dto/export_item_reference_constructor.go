package dto

func NewExportItemReference(localId, itemKind, itemId string) ExportItemReference {
	return ExportItemReference{LocalId: localId, ItemKind: itemKind, ItemId: itemId}
}
