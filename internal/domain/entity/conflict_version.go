package entity

type ConflictVersion struct {
	Request     *Request
	Environment *Environment
	Text        string
}

type ConflictVersions struct {
	Base   *ConflictVersion
	Mine   *ConflictVersion
	Theirs *ConflictVersion
}
