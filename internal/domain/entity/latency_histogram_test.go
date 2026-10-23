package entity

import (
	"math"
	"math/rand"
	"slices"
	"testing"
	"time"
)

func exactPercentile(sorted []time.Duration, percent int) time.Duration {
	rank := max(1, (len(sorted)*percent+99)/100)
	return sorted[rank-1]
}

func TestLatencyHistogramPercentilesAreWithinOnePercent(t *testing.T) {
	random := rand.New(rand.NewSource(17))
	for _, scale := range []time.Duration{time.Microsecond, time.Millisecond, 50 * time.Millisecond, time.Second, time.Minute} {
		histogram := NewLatencyHistogram()
		samples := make([]time.Duration, 0, 5000)
		for range 5000 {
			sample := time.Duration(random.ExpFloat64()*float64(scale)) + 200*time.Microsecond
			samples = append(samples, sample)
			histogram.Record(sample)
		}
		slices.Sort(samples)
		for _, percent := range []int{50, 90, 95, 99, 100} {
			exact := exactPercentile(samples, percent)
			got := histogram.Percentile(percent)
			if relative := math.Abs(float64(got-exact)) / float64(exact); relative > 0.01 {
				t.Errorf("scale %s p%d = %s, exact %s (%.4f)", scale, percent, got, exact, relative)
			}
		}
	}
}

func TestLatencyHistogramKeepsExactMinMaxSumAndCount(t *testing.T) {
	histogram := NewLatencyHistogram()
	for _, sample := range []time.Duration{30 * time.Millisecond, 10 * time.Millisecond, 20 * time.Millisecond} {
		histogram.Record(sample)
	}
	if histogram.Min != 10*time.Millisecond || histogram.Max != 30*time.Millisecond || histogram.Sum != 60*time.Millisecond || histogram.Count != 3 {
		t.Errorf("histogram = min %s max %s sum %s count %d", histogram.Min, histogram.Max, histogram.Sum, histogram.Count)
	}
	if histogram.Mean() != 20*time.Millisecond {
		t.Errorf("mean = %s", histogram.Mean())
	}
	if histogram.Percentile(100) != 30*time.Millisecond || histogram.Percentile(1) != 10*time.Millisecond {
		t.Errorf("p100 = %s, p1 = %s", histogram.Percentile(100), histogram.Percentile(1))
	}
}

func TestLatencyHistogramIsEmptyWithoutSamples(t *testing.T) {
	histogram := NewLatencyHistogram()
	if histogram.Percentile(99) != 0 || histogram.Mean() != 0 {
		t.Errorf("empty histogram = %+v", histogram)
	}
}

func TestLatencyHistogramClampsSamplesBeyondItsRange(t *testing.T) {
	histogram := NewLatencyHistogram()
	histogram.Record(time.Hour)
	if histogram.Percentile(50) != time.Hour || histogram.Max != time.Hour {
		t.Errorf("p50 = %s", histogram.Percentile(50))
	}
}
