package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type AuthDigest struct {
	Username            string `yaml:"username" json:"username"`
	Password            string `yaml:"password" json:"password"`
	Algorithm           string `yaml:"algorithm" json:"algorithm"`
	Qop                 string `yaml:"qop" json:"qop"`
	Realm               string `yaml:"realm" json:"realm"`
	Nonce               string `yaml:"nonce" json:"nonce"`
	Opaque              string `yaml:"opaque" json:"opaque"`
	NonceCount          string `yaml:"nonceCount" json:"nonceCount"`
	ClientNonce         string `yaml:"clientNonce" json:"clientNonce"`
	DisableRetryRequest bool   `yaml:"disableRetryRequest" json:"disableRetryRequest"`
}

func NewAuthDigest(digest entity.DigestAuth) AuthDigest {
	return AuthDigest{
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

func (a *AuthDigest) fields() []yamlnode.Field {
	digest := a.ToEntity()
	return []yamlnode.Field{
		yamlnode.String("username", digest.Username),
		yamlnode.String("password", digest.Password),
		yamlnode.String("algorithm", digest.Algorithm.String()),
		yamlnode.String("qop", digest.Qop.String()),
		yamlnode.String("realm", digest.Realm),
		yamlnode.String("nonce", digest.Nonce),
		yamlnode.String("opaque", digest.Opaque),
		yamlnode.String("nonceCount", digest.NonceCount),
		yamlnode.String("clientNonce", digest.ClientNonce),
		yamlnode.Bool("disableRetryRequest", digest.DisableRetryRequest),
	}
}
