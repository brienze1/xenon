package entity

import "strings"

type ReleaseChecksums map[string]string

func (r ReleaseChecksums) Lists(name string) bool {
	_, listed := r[name]
	return listed
}

func (r ReleaseChecksums) Matches(name, digest string) bool {
	expected, listed := r[name]
	return listed && digest != "" && strings.EqualFold(expected, digest)
}
