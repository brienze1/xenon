package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

// ValidateAppSettings returns the first failing field in order: script timeout, user cap, AI cap, port.
func ValidateAppSettings(settings entity.AppSettings) error {
	switch {
	case !settings.HasValidScriptTimeout():
		return errs.BadRequestError("the script timeout must be between 1000 and 60000 ms", "SCR-01400",
			errs.ErrorDetails{Attribute: "script_timeout_ms", Messages: []string{"must be between 1000 and 60000"}})
	case !settings.HasValidHistoryUserCap():
		return errs.BadRequestError("the History cap for your sends must be between 100 and 20000", "SET-02400",
			errs.ErrorDetails{Attribute: "history_user_cap", Messages: []string{"must be between 100 and 20000"}})
	case !settings.HasValidHistoryAICap():
		return errs.BadRequestError("the History cap for AI activity must be between 100 and 20000", "SET-03400",
			errs.ErrorDetails{Attribute: "history_ai_cap", Messages: []string{"must be between 100 and 20000"}})
	case !settings.HasValidAIPort():
		return errs.BadRequestError("the AI access port must be between 1024 and 65535", "SET-04400",
			errs.ErrorDetails{Attribute: "ai_access.port", Messages: []string{"must be between 1024 and 65535"}})
	default:
		return nil
	}
}
