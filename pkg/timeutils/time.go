package timeutils

import "time"

func Now() time.Time {
	return getTimeConfig().provider()().UTC()
}
