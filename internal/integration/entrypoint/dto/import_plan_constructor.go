package dto

func NewImportPlan(sourceId, applyToAll, passphrase string, items []ImportPlanItem) ImportPlan {
	return ImportPlan{SourceId: sourceId, ApplyToAll: applyToAll, Passphrase: passphrase, Items: items}
}
