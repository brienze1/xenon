package entity

import "github.com/brienze1/xenon/internal/domain/enums"

// ScriptMember is one entry of the sandbox surface offered to the script editor's autocomplete.
type ScriptMember struct {
	Kind        enums.ScriptMemberKind
	Path        string
	Signature   string
	Description string
}
