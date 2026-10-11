package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	verifierAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	verifierLength   = 64
	stateBytes       = 16
)

func GeneratePkce(config entity.OAuth2Config) entity.Pkce {
	verifier := config.CodeVerifier
	if verifier == "" {
		verifier = randomVerifier()
	}
	if config.UsesPlainChallenge() {
		return entity.Pkce{Verifier: verifier, Challenge: verifier, Method: entity.PkcePlain}
	}
	sum := sha256.Sum256([]byte(verifier))
	return entity.Pkce{Verifier: verifier, Challenge: base64.RawURLEncoding.EncodeToString(sum[:]), Method: entity.PkceS256}
}

func GenerateOAuthState(config entity.OAuth2Config) string {
	if config.State != "" {
		return config.State
	}
	random := make([]byte, stateBytes)
	_, _ = rand.Read(random)
	return hex.EncodeToString(random)
}

func randomVerifier() string {
	random := make([]byte, verifierLength)
	_, _ = rand.Read(random)
	verifier := make([]byte, verifierLength)
	for index, value := range random {
		verifier[index] = verifierAlphabet[int(value)%len(verifierAlphabet)]
	}
	return string(verifier)
}
