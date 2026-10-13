package entity

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

type PostmanDump struct {
	Collections  []ImportDocument
	Environments []ImportDocument
	Skipped      []PostmanDumpPart
}

func (p PostmanDump) SkippedWarnings() []ImportWarning {
	warnings := make([]ImportWarning, 0, len(p.Skipped))
	for _, part := range p.Skipped {
		message := "The dump's " + part.Name + " part (" + strconv.Itoa(part.Files) + " " + fileNoun(part.Files) + ") is not a collection or an environment and was skipped"
		warnings = append(warnings, NewImportWarning(warningcode.SkippedPart, message, "", part.Name))
	}
	return warnings
}

func fileNoun(count int) string {
	if count == 1 {
		return "file"
	}
	return "files"
}
