package errors

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const actionAttribute = "action"

func RejectUnknownAIAction(tool, action string, actions []string) error {
	return errs.BadRequestError("Unknown action \""+action+"\" for "+tool+" — use one of: "+strings.Join(actions, ", "), entity.AIInvalidArgumentsCode,
		errs.ErrorDetails{Attribute: actionAttribute, Messages: actions})
}
