package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const clashesAttribute = "clashes"

func RejectUnresolvedClashes(clashes []entity.Clash) error {
	messages := make([]string, 0, len(clashes))
	for _, clash := range clashes {
		messages = append(messages, clash.Description())
	}
	return errs.ConflictError("choose Replace, Keep both or Skip for the name clashes before importing", "IMP-04409", errs.ErrorDetails{Attribute: clashesAttribute, Messages: messages})
}
