package timeutils

import (
	"testing"
	"time"
)

func TestNowUsesConfiguredProviderInUtc(t *testing.T) {
	fixed := time.Date(2026, 10, 9, 12, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
	SetCurrentTimeProvider(func() time.Time { return fixed })
	t.Cleanup(func() { SetCurrentTimeProvider(time.Now) })
	got := Now()
	if !got.Equal(fixed) || got.Location() != time.UTC {
		t.Errorf("Now() = %v, want %v in UTC", got, fixed)
	}
}

func TestNowDefaultsToRealTime(t *testing.T) {
	SetCurrentTimeProvider(time.Now)
	before := time.Now().Add(-time.Second)
	if got := Now(); got.Before(before) || got.Location() != time.UTC {
		t.Errorf("Now() = %v", got)
	}
}
