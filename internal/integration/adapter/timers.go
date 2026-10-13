package adapter

import "time"

type Timers interface {
	AfterFunc(delay time.Duration, run func()) (stop func() bool)
}
