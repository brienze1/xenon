package entity

import (
	"math/bits"
	"time"
)

const (
	latencyPrecisionBits = 7
	latencySubBuckets    = 1 << latencyPrecisionBits
	latencyMaxMicros     = 1<<30 - 1
	latencyBuckets       = (31 - latencyPrecisionBits) * latencySubBuckets
)

// LatencyHistogram is a log-linear histogram of microseconds, within 0.4% of each sample up to 17 minutes.
type LatencyHistogram struct {
	Buckets []int64
	Min     time.Duration
	Max     time.Duration
	Sum     time.Duration
	Count   int
}

func NewLatencyHistogram() LatencyHistogram {
	return LatencyHistogram{Buckets: make([]int64, latencyBuckets)}
}

func (l *LatencyHistogram) Record(latency time.Duration) {
	latency = max(latency, 0)
	if l.Count == 0 || latency < l.Min {
		l.Min = latency
	}
	if latency > l.Max {
		l.Max = latency
	}
	l.Sum += latency
	l.Count++
	l.Buckets[latencyBucketOf(latency)]++
}

func (l *LatencyHistogram) Mean() time.Duration {
	if l.Count == 0 {
		return 0
	}
	return l.Sum / time.Duration(l.Count)
}

// Percentile is the nearest-rank percentile, kept within the exact min and max.
func (l *LatencyHistogram) Percentile(percent int) time.Duration {
	if l.Count == 0 {
		return 0
	}
	rank := max(1, int64((l.Count*percent+99)/100))
	switch rank {
	case 1:
		return l.Min
	case int64(l.Count):
		return l.Max
	}
	var seen int64
	for index, count := range l.Buckets {
		seen += count
		if seen >= rank {
			return min(max(latencyBucketValue(index), l.Min), l.Max)
		}
	}
	return l.Max
}

func latencyBucketOf(latency time.Duration) int {
	micros := min(uint64(latency.Microseconds()), latencyMaxMicros)
	if micros < latencySubBuckets {
		return int(micros)
	}
	magnitude := bits.Len64(micros) - 1 - latencyPrecisionBits
	return magnitude*latencySubBuckets + int(micros>>magnitude)
}

func latencyBucketValue(index int) time.Duration {
	if index < latencySubBuckets {
		return time.Duration(index) * time.Microsecond
	}
	magnitude := index/latencySubBuckets - 1
	lower := uint64(index%latencySubBuckets+latencySubBuckets) << magnitude
	width := uint64(1) << magnitude
	return time.Duration(lower+(width-1)/2) * time.Microsecond
}
