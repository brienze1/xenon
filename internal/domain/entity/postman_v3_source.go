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
	Skipped      []PostmanDumpPart
}

func NewPostmanV3Source() PostmanV3Source {
	return PostmanV3Source{Collections: []ImportCollection{}, Environments: []ImportVariables{}, Skipped: []PostmanDumpPart{}}
}

func (p PostmanV3Source) WithRequest(request ImportCollection) PostmanV3Source {
	p.Request = &request
	return p
}

func (p PostmanV3Source) WithCollection(collection ImportCollection) PostmanV3Source {
	p.Collections = append(p.Collections, collection)
	return p
}

func (p PostmanV3Source) WithEnvironment(environment ImportVariables) PostmanV3Source {
	p.Environments = append(p.Environments, environment)
	return p
}

func (p PostmanV3Source) WithGlobals(globals ImportVariables) PostmanV3Source {
	p.Globals = &globals
	return p
}

func (p PostmanV3Source) WithSkipped(skipped []PostmanDumpPart) PostmanV3Source {
	p.Skipped = skipped
	return p
}

func (p PostmanV3Source) HasRequest() bool {
	return p.Request != nil
}

func (p PostmanV3Source) HasGlobals() bool {
	return p.Globals != nil
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
	return map[string]any{"collections": len(p.Collections), "environments": len(p.Environments), "globals": p.HasGlobals(), "request": p.HasRequest(), "skipped": len(p.Skipped)}
}
