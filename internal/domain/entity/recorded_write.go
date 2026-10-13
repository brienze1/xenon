package entity

import (
	"time"

	"github.com/brienze1/xenon/pkg/timeutils"
)

const ownWriteWindow = 2 * time.Second

type RecordedWrite struct {
	At       time.Time
	Revision string
}

func NewRecordedWrite(revision string) RecordedWrite {
	return RecordedWrite{At: timeutils.Now(), Revision: revision}
}

func (r RecordedWrite) Echoes(revision string) bool {
	return r.Revision == revision && timeutils.Now().Sub(r.At) <= ownWriteWindow
}
