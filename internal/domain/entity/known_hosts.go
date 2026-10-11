package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/hostkeystatus"
)

type KnownHosts struct {
	Keys []HostKey
}

func (k KnownHosts) StatusOf(presented HostKey) enums.HostKeyStatus {
	known := k.KeysFor(presented.Host)
	if len(known) == 0 {
		return hostkeystatus.Unrecognized
	}
	for _, key := range known {
		if key.SameKey(presented) {
			return hostkeystatus.Trusted
		}
	}
	return hostkeystatus.Changed
}

func (k KnownHosts) KeysFor(host string) []HostKey {
	keys := make([]HostKey, 0)
	for _, key := range k.Keys {
		if key.Host == host {
			keys = append(keys, key)
		}
	}
	return keys
}
