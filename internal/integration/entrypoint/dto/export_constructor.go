package dto

func NewExport(localId, itemKind, itemId, format, shape, targetPath string, currentValuesAsInitial bool) Export {
	return Export{
		LocalId:                localId,
		ItemKind:               itemKind,
		ItemId:                 itemId,
		Format:                 format,
		Shape:                  shape,
		TargetPath:             targetPath,
		CurrentValuesAsInitial: currentValuesAsInitial,
	}
}
