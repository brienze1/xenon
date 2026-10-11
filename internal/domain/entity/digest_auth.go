package entity

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
)

type DigestAuth struct {
	Algorithm           enums.DigestAlgorithm
	Qop                 enums.DigestQop
	Username            string
	Password            string
	Realm               string
	Nonce               string
	Opaque              string
	NonceCount          string
	ClientNonce         string
	DisableRetryRequest bool
}

func DefaultDigestAuth() DigestAuth {
	return DigestAuth{Algorithm: digestalgorithm.MD5, Qop: digestqop.None}
}

func (d DigestAuth) IsBlank() bool {
	return d.Username == "" && d.Password == "" && d.Realm == "" && d.Nonce == "" && d.Opaque == ""
}

func (d DigestAuth) RetriesChallenge() bool {
	return !d.DisableRetryRequest
}

// TypedState is the challenge the user typed in the advanced fields, if any.
func (d DigestAuth) TypedState() (DigestState, bool) {
	if d.Nonce == "" {
		return DigestState{}, false
	}
	count, err := strconv.ParseInt(d.NonceCount, 16, 64)
	if err != nil {
		count = 0
	}
	return DigestState{Realm: d.Realm, Nonce: d.Nonce, Opaque: d.Opaque, Algorithm: d.Algorithm, Qop: d.Qop, NonceCount: int(count)}, true
}

func (d DigestAuth) MapText(transform func(string) string) DigestAuth {
	d.Username = transform(d.Username)
	d.Password = transform(d.Password)
	d.Realm = transform(d.Realm)
	d.Nonce = transform(d.Nonce)
	d.Opaque = transform(d.Opaque)
	d.NonceCount = transform(d.NonceCount)
	d.ClientNonce = transform(d.ClientNonce)
	return d
}

func (d DigestAuth) TextFields(prefix string) []TextField {
	return []TextField{NewTextField(fieldName(prefix, "digest", "username"), d.Username), NewTextField(fieldName(prefix, "digest", "password"), d.Password)}
}

func (d DigestAuth) LogValue() any {
	return map[string]any{"username": d.Username, "algorithm": d.Algorithm, "qop": d.Qop, "hasPassword": d.Password != ""}
}
