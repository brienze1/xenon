package entity

import "time"

type Timings struct {
	Dns       time.Duration
	Connect   time.Duration
	Tls       time.Duration
	FirstByte time.Duration
	Download  time.Duration
	Total     time.Duration
}
