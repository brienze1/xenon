package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type LoadProfileOverrides struct {
	Kind         *enums.LoadProfileKind
	InitialUsers *int
	RampSeconds  *int
	BaseUsers    *int
	SpikeSeconds *int
	PeakSeconds  *int
}
