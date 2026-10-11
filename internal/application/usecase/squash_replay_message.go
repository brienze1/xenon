package usecase

import (
	"fmt"
	"strings"
)

const syncSubjectPrefix = "Sync"

func SquashReplayMessage(messages []string) string {
	originals := make([]string, 0, len(messages))
	for _, message := range messages {
		if trimmed := strings.TrimSpace(message); trimmed != "" {
			originals = append(originals, trimmed)
		}
	}
	if len(originals) == 0 {
		return syncSubjectPrefix
	}
	subject := syncSubjectPrefix + ": " + strings.SplitN(originals[0], "\n", 2)[0]
	if len(originals) > 1 {
		subject += fmt.Sprintf(" (+%d more)", len(originals)-1)
	}
	body := make([]string, 0, len(originals))
	for _, original := range originals {
		body = append(body, "- "+strings.ReplaceAll(original, "\n", "\n  "))
	}
	return subject + "\n\n" + strings.Join(body, "\n")
}
