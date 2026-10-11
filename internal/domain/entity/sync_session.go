package entity

import "time"

type SyncSession struct {
	StartedAt  time.Time
	ID         string
	LocalHead  string
	RemoteHead string
	Base       string
	Conflicts  []ConflictItem
}
