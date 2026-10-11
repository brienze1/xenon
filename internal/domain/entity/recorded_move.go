package entity

import (
	"time"

	"github.com/brienze1/xenon/pkg/timeutils"
)

type RecordedMove struct {
	At     time.Time
	Source string
	Target string
}

func NewRecordedMove(source, target string) RecordedMove {
	return RecordedMove{At: timeutils.Now(), Source: source, Target: target}
}

func (r RecordedMove) IsRecent() bool {
	return timeutils.Now().Sub(r.At) <= ownWriteWindow
}

func (r RecordedMove) Covers(path string) bool {
	return r.IsRecent() && (isWithin(path, r.Source) || isWithin(path, r.Target))
}
