package entity

import "testing"

func TestInstallationAcceptsOnlyTheExpectedPackageEntry(t *testing.T) {
	bundle := Installation{Platform: "darwin-arm64"}
	plain := Installation{Platform: "linux-amd64"}
	cases := []struct {
		installation Installation
		name         string
		isDir        bool
		isRegular    bool
		want         bool
	}{
		{bundle, "Xenon.app", true, false, true},
		{bundle, "Xenon.app", false, true, false},
		{bundle, "xenon", false, true, false},
		{plain, "xenon", false, true, true},
		{plain, "xenon.exe", false, true, true},
		{plain, "xenon", true, false, false},
		{plain, "other", false, true, false},
	}
	for _, c := range cases {
		if got := c.installation.AcceptsPackageEntry(c.name, c.isDir, c.isRegular); got != c.want {
			t.Errorf("%s.AcceptsPackageEntry(%q, %v, %v) = %v; want %v", c.installation.Platform, c.name, c.isDir, c.isRegular, got, c.want)
		}
	}
}
