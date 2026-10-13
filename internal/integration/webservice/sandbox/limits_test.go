package sandbox

import "testing"

func TestExceedsGrowth(t *testing.T) {
	cases := []struct {
		current, baseline, growth uint64
		want                      bool
	}{
		{current: 100, baseline: 40, growth: 60, want: false},
		{current: 101, baseline: 40, growth: 60, want: true},
		{current: 99, baseline: 40, growth: 60, want: false},
		{current: 10, baseline: 40, growth: 60, want: false},
	}
	for _, tc := range cases {
		if got := exceedsGrowth(tc.current, tc.baseline, tc.growth); got != tc.want {
			t.Errorf("exceedsGrowth(%d, %d, %d) = %v", tc.current, tc.baseline, tc.growth, got)
		}
	}
}

func TestHeapBytesReportsTheLiveHeap(t *testing.T) {
	if heapBytes() == 0 {
		t.Fatal("heap metric is not available")
	}
}
