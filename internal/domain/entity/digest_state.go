package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type DigestState struct {
	Algorithm  enums.DigestAlgorithm
	Qop        enums.DigestQop
	Realm      string
	Nonce      string
	Opaque     string
	NonceCount int
}

func (d DigestState) IsEmpty() bool {
	return d.Nonce == ""
}

func (d DigestState) Next() DigestState {
	d.NonceCount++
	return d
}

func (d DigestState) LogValue() any {
	return map[string]any{"realm": d.Realm, "algorithm": d.Algorithm, "qop": d.Qop, "nonceCount": d.NonceCount}
}
