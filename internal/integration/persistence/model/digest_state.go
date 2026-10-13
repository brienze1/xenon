package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
)

type DigestState struct {
	Algorithm  string `json:"algorithm"`
	Qop        string `json:"qop"`
	Realm      string `json:"realm"`
	Nonce      string `json:"nonce"`
	Opaque     string `json:"opaque"`
	NonceCount int    `json:"nonce_count"`
}

func NewDigestState(state entity.DigestState) DigestState {
	return DigestState{Algorithm: state.Algorithm.String(), Qop: state.Qop.String(), Realm: state.Realm, Nonce: state.Nonce, Opaque: state.Opaque, NonceCount: state.NonceCount}
}

func (d *DigestState) ToEntity() entity.DigestState {
	return entity.DigestState{Algorithm: digestalgorithm.GetDigestAlgorithm(d.Algorithm), Qop: digestqop.GetDigestQop(d.Qop), Realm: d.Realm, Nonce: d.Nonce, Opaque: d.Opaque, NonceCount: d.NonceCount}
}
