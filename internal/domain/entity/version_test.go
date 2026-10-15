package entity

import "testing"

func TestParseVersionAcceptsReleaseVersions(t *testing.T) {
	cases := map[string]Version{"0.1.7": {0, 1, 7}, "v0.1.10": {0, 1, 10}, " v1.2.3 ": {1, 2, 3}}
	for text, want := range cases {
		got, err := ParseVersion(text)
		if err != nil || got != want {
			t.Errorf("ParseVersion(%q) = %v, %v; want %v", text, got, err, want)
		}
	}
}

func TestParseVersionRejectsAnythingElse(t *testing.T) {
	for _, text := range []string{"", "nightly", "0.1", "0.1.7-beta", "0.1.x", "1.2.3.4", "v-1.2.3", "01.2.3", "1.+2.3"} {
		if _, err := ParseVersion(text); err == nil {
			t.Errorf("ParseVersion(%q) should fail", text)
		}
	}
}

func TestNewerThanComparesNumerically(t *testing.T) {
	newer := []([2]Version){
		{{0, 1, 10}, {0, 1, 9}},
		{{0, 2, 0}, {0, 1, 99}},
		{{1, 0, 0}, {0, 9, 9}},
	}
	for _, pair := range newer {
		if !pair[0].NewerThan(pair[1]) || pair[1].NewerThan(pair[0]) {
			t.Errorf("%s should be newer than %s", pair[0], pair[1])
		}
	}
	if (Version{0, 1, 7}).NewerThan(Version{0, 1, 7}) {
		t.Error("a version is not newer than itself")
	}
}
