package entity

type ConflictVersions struct {
	Base   *ConflictVersion
	Mine   *ConflictVersion
	Theirs *ConflictVersion
}
