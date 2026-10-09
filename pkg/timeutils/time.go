package timeutils

import "time"

func Now() time.Time {
	return GetTimeConfig().provider()().UTC()
}
