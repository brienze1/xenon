package entity

import "testing"

func TestAILoopbackHostAcceptsOnlyTheLoopbackNamesOnThePort(t *testing.T) {
	cases := map[string]bool{
		"127.0.0.1:7777":    true,
		"localhost:7777":    true,
		"127.0.0.1:8888":    false,
		"127.0.0.1":         false,
		"evil.example:7777": false,
	}
	for host, want := range cases {
		if got := IsAILoopbackHost(host, 7777); got != want {
			t.Errorf("IsAILoopbackHost(%q) = %v, want %v", host, got, want)
		}
	}
}

func TestAILoopbackOriginAcceptsTheLoopbackOriginsWithOrWithoutThePort(t *testing.T) {
	cases := map[string]bool{
		"http://127.0.0.1":      true,
		"http://127.0.0.1:7777": true,
		"http://localhost":      true,
		"http://localhost:7777": true,
		"http://localhost:8888": false,
		"https://localhost":     false,
		"null":                  false,
		"":                      false,
	}
	for origin, want := range cases {
		if got := IsAILoopbackOrigin(origin, 7777); got != want {
			t.Errorf("IsAILoopbackOrigin(%q) = %v, want %v", origin, got, want)
		}
	}
}

func TestAIAccessStatusDiffersOnlyWhenAFieldChanges(t *testing.T) {
	running := RunningAIAccessStatus(7777, 1)
	if running.Differs(RunningAIAccessStatus(7777, 1)) {
		t.Error("equal statuses must not differ")
	}
	if !running.Differs(RunningAIAccessStatus(7777, 2)) || !running.Differs(OffAIAccessStatus(7777)) {
		t.Error("changed statuses must differ")
	}
}
