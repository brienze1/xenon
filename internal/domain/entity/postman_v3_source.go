package entity

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

type PostmanV3Source struct {
	Request      *ImportCollection
	Globals      *ImportVariables
	Collections  []ImportCollection
	Environments []ImportVariables
	Warnings     []ImportWarning
	Skipped      []PostmanDumpPart
}

func NewPostmanV3Source() PostmanV3Source {
	return PostmanV3Source{Collections: []ImportCollection{}, Environments: []ImportVariables{}, Warnings: []ImportWarning{}, Skipped: []PostmanDumpPart{}}
}

func (p PostmanV3Source) SkippedWarnings() []ImportWarning {
	warnings := make([]ImportWarning, 0, len(p.Skipped))
	for _, part := range p.Skipped {
		message := "The workspace's " + part.Name + " part (" + strconv.Itoa(part.Files) + " " + fileNoun(part.Files) + ") is not a collection, an environment or globals and was skipped"
		warnings = append(warnings, NewImportWarning(warningcode.SkippedPart, message, "", part.Name))
	}
	return warnings
}

func (p PostmanV3Source) LogValue() any {
	return map[string]any{"collections": len(p.Collections), "environments": len(p.Environments), "globals": p.Globals != nil, "request": p.Request != nil, "warnings": len(p.Warnings), "skipped": len(p.Skipped)}
}
