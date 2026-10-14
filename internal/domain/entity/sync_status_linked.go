package entity

import "github.com/brienze1/xenon/internal/domain/enums/syncstate"

func (s SyncStatus) IsLinked() bool {
	return s.State != syncstate.NotLinked
}
