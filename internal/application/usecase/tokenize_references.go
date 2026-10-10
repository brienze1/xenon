package usecase

import (
	"regexp"

	"github.com/brienze1/xenon/internal/domain/entity"
)

var variableReference = regexp.MustCompile(`\{\{([^{}]*?)\}\}`)

func TokenizeReferences(field, text string) []entity.Reference {
	matches := variableReference.FindAllStringSubmatchIndex(text, -1)
	references := make([]entity.Reference, 0, len(matches))
	for _, match := range matches {
		references = append(references, entity.Reference{Name: text[match[2]:match[3]], Field: field, Start: match[0], End: match[1]})
	}
	return references
}
