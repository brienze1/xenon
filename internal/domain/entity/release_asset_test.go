package entity

import "testing"

func TestReleaseAssetExpectedSizePrefersTheReportedLength(t *testing.T) {
	asset := ReleaseAsset{Name: "checksums.txt", Size: 42}
	cases := map[int64]int64{100: 100, 0: 42, -1: 42}
	for reported, want := range cases {
		if got := asset.ExpectedSize(reported); got != want {
			t.Errorf("ExpectedSize(%d) = %d; want %d", reported, got, want)
		}
	}
	if !asset.IsNamed("checksums.txt") || asset.IsNamed("other.txt") {
		t.Errorf("IsNamed should match only the asset name")
	}
}
