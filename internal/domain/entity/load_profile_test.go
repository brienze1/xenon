package entity

import (
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/loadprofilekind"
)

func TestLoadProfileActiveUsersAt(t *testing.T) {
	cases := []struct {
		name     string
		profile  LoadProfile
		at       time.Duration
		users    int
		duration int
		want     int
	}{
		{"fixed at the start", LoadProfile{Kind: loadprofilekind.Fixed}, 0, 10, 60, 10},
		{"fixed at the end", LoadProfile{Kind: loadprofilekind.Fixed}, 59 * time.Second, 10, 60, 10},
		{"ramp up starts at the initial users", LoadProfile{Kind: loadprofilekind.RampUp, InitialUsers: 1, RampSeconds: 10}, 0, 11, 60, 1},
		{"ramp up is linear", LoadProfile{Kind: loadprofilekind.RampUp, InitialUsers: 1, RampSeconds: 10}, 5 * time.Second, 11, 60, 6},
		{"ramp up holds the maximum after the ramp", LoadProfile{Kind: loadprofilekind.RampUp, InitialUsers: 1, RampSeconds: 10}, 10 * time.Second, 11, 60, 11},
		{"spike holds the base before the window", LoadProfile{Kind: loadprofilekind.Spike, BaseUsers: 2, SpikeSeconds: 10}, 24 * time.Second, 20, 60, 2},
		{"spike is centred", LoadProfile{Kind: loadprofilekind.Spike, BaseUsers: 2, SpikeSeconds: 10}, 25 * time.Second, 20, 60, 20},
		{"spike ends after its duration", LoadProfile{Kind: loadprofilekind.Spike, BaseUsers: 2, SpikeSeconds: 10}, 35 * time.Second, 20, 60, 2},
		{"peak starts at the base", LoadProfile{Kind: loadprofilekind.Peak, BaseUsers: 2, PeakSeconds: 20}, 0, 22, 60, 2},
		{"peak ramps up over half the rest", LoadProfile{Kind: loadprofilekind.Peak, BaseUsers: 2, PeakSeconds: 20}, 10 * time.Second, 22, 60, 12},
		{"peak holds the maximum", LoadProfile{Kind: loadprofilekind.Peak, BaseUsers: 2, PeakSeconds: 20}, 30 * time.Second, 22, 60, 22},
		{"peak ramps down over the other half", LoadProfile{Kind: loadprofilekind.Peak, BaseUsers: 2, PeakSeconds: 20}, 50 * time.Second, 22, 60, 12},
		{"peak ends at the base", LoadProfile{Kind: loadprofilekind.Peak, BaseUsers: 2, PeakSeconds: 20}, 60 * time.Second, 22, 60, 2},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.profile.ActiveUsersAt(testCase.at, testCase.users, testCase.duration); got != testCase.want {
				t.Errorf("ActiveUsersAt(%s) = %d, want %d", testCase.at, got, testCase.want)
			}
		})
	}
}

func TestLoadProfileWithDefaultsFillsOnlyTheZeroParameters(t *testing.T) {
	cases := []struct {
		name    string
		profile LoadProfile
		want    LoadProfile
	}{
		{"ramp up", LoadProfile{Kind: loadprofilekind.RampUp}, LoadProfile{Kind: loadprofilekind.RampUp, InitialUsers: 1, RampSeconds: 150}},
		{"spike", LoadProfile{Kind: loadprofilekind.Spike}, LoadProfile{Kind: loadprofilekind.Spike, BaseUsers: 5, SpikeSeconds: 60}},
		{"peak", LoadProfile{Kind: loadprofilekind.Peak}, LoadProfile{Kind: loadprofilekind.Peak, BaseUsers: 5, PeakSeconds: 300}},
		{"explicit values are kept", LoadProfile{Kind: loadprofilekind.Spike, BaseUsers: 7, SpikeSeconds: 3}, LoadProfile{Kind: loadprofilekind.Spike, BaseUsers: 7, SpikeSeconds: 3}},
		{"fixed has no parameters", LoadProfile{Kind: loadprofilekind.Fixed}, LoadProfile{Kind: loadprofilekind.Fixed}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.profile.WithDefaults(50, 600); got != testCase.want {
				t.Errorf("WithDefaults = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func TestLoadProfileDefaultsAreAtLeastOne(t *testing.T) {
	spike := LoadProfile{Kind: loadprofilekind.Spike}.WithDefaults(3, 5)
	if spike.BaseUsers != 1 || spike.SpikeSeconds != 1 {
		t.Errorf("spike = %+v", spike)
	}
	ramp := LoadProfile{Kind: loadprofilekind.RampUp}.WithDefaults(3, 2)
	if ramp.RampSeconds != 1 {
		t.Errorf("ramp = %+v", ramp)
	}
}
