package entity

import "github.com/brienze1/xenon/internal/domain/enums"

// HistoryEnvironment is the environment active for a send, as it was then.
type HistoryEnvironment struct {
	ID    string
	Name  string
	Color enums.EnvColor
}

// NewHistoryEnvironment snapshots the active environment, or is nil when none is active.
func NewHistoryEnvironment(active *Environment) *HistoryEnvironment {
	if active == nil {
		return nil
	}
	return &HistoryEnvironment{ID: active.ID, Name: active.Name, Color: active.Color}
}
