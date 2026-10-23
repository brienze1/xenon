package entity

import "github.com/brienze1/xenon/internal/domain/enums"

// HistoryEnvironment is the environment active for a send, as it was then.
type HistoryEnvironment struct {
	ID    string
	Name  string
	Color enums.EnvColor
}
