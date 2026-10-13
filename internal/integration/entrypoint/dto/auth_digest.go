package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
)

type AuthDigest struct {
	Username            string `json:"username"`
	Password            string `json:"password"`
	Algorithm           string `json:"algorithm" validate:"omitempty,oneof=MD5 MD5-sess SHA-256 SHA-256-sess SHA-512-256 SHA-512-256-sess"`
	Qop                 string `json:"qop" validate:"omitempty,oneof=auth auth-int"`
	Realm               string `json:"realm"`
	Nonce               string `json:"nonce"`
	Opaque              string `json:"opaque"`
	NonceCount          string `json:"nonce_count"`
	ClientNonce         string `json:"client_nonce"`
	DisableRetryRequest bool   `json:"disable_retry_request"`
}

func NewAuthDigest(digest entity.DigestAuth) *AuthDigest {
	return &AuthDigest{
		Username:            digest.Username,
		Password:            digest.Password,
		Algorithm:           digest.Algorithm.String(),
		Qop:                 digest.Qop.String(),
		Realm:               digest.Realm,
		Nonce:               digest.Nonce,
		Opaque:              digest.Opaque,
		NonceCount:          digest.NonceCount,
		ClientNonce:         digest.ClientNonce,
		DisableRetryRequest: digest.DisableRetryRequest,
	}
}

func (a *AuthDigest) ToEntity() entity.DigestAuth {
	if a == nil {
		return entity.DigestAuth{}
	}
	digest := entity.DefaultDigestAuth()
	if algorithm := digestalgorithm.GetDigestAlgorithm(a.Algorithm); algorithm != digestalgorithm.Unknown {
		digest.Algorithm = algorithm
	}
	if qop := digestqop.GetDigestQop(a.Qop); qop != digestqop.Unknown {
		digest.Qop = qop
	}
	digest.Username = a.Username
	digest.Password = a.Password
	digest.Realm = a.Realm
	digest.Nonce = a.Nonce
	digest.Opaque = a.Opaque
	digest.NonceCount = a.NonceCount
	digest.ClientNonce = a.ClientNonce
	digest.DisableRetryRequest = a.DisableRetryRequest
	return digest
}
