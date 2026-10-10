package entity

type ImportOutcome struct {
	Failure  *ImportFailure
	Item     ImportItemResult
	Warnings []ImportWarning
	PmUsages []PmUsage
	Created  ImportCounts
	Replaced bool
	Skipped  bool
}
