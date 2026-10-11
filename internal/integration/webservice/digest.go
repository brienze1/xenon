package webservice

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"io"
	"net/url"
	"strings"

	"github.com/icholy/digest"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	sessionSuffix     = "-sess"
	clientNonceLength = 8
)

type digestWebservice struct{}

func NewDigestWebservice() adapter.DigestWebservice {
	return &digestWebservice{}
}

func (d *digestWebservice) ReadChallenge(ctx context.Context, challenge string) (*entity.DigestState, error) {
	logger.Debug(ctx, "Started", map[string]any{"challenge": challenge})

	parsed, err := digest.ParseChallenge(challenge)
	if err != nil {
		return nil, errs.BadRequestError("the server's Digest challenge could not be read: "+err.Error(), "DIG-01400")
	}
	state := entity.DigestState{Realm: parsed.Realm, Nonce: parsed.Nonce, Opaque: parsed.Opaque, Algorithm: challengeAlgorithm(parsed.Algorithm), Qop: challengeQop(parsed.QOP)}

	logger.Debug(ctx, "Finished", state)
	return &state, nil
}

func (d *digestWebservice) Answer(ctx context.Context, request entity.EffectiveRequest, auth entity.DigestAuth, state entity.DigestState) (string, error) {
	logger.Debug(ctx, "Started", map[string]any{"request": request, "auth": auth, "state": state})

	algorithm := state.Algorithm
	if algorithm == "" {
		algorithm = auth.Algorithm
	}
	base := strings.TrimSuffix(algorithm.String(), sessionSuffix)
	clientNonce := auth.ClientNonce
	if clientNonce == "" {
		clientNonce = newClientNonce()
	}
	challenge := &digest.Challenge{Realm: state.Realm, Nonce: state.Nonce, Opaque: state.Opaque, Algorithm: base}
	if state.Qop != digestqop.None {
		challenge.QOP = []string{state.Qop.String()}
	}
	options := digest.Options{Method: request.Method, URI: requestUri(request.URL), Count: state.NonceCount, Username: auth.Username, Password: auth.Password, Cnonce: clientNonce}
	options.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(payload(request.Body))), nil }
	if strings.HasSuffix(strings.ToLower(algorithm.String()), sessionSuffix) {
		newHash := digestHash(base)
		options.A1 = hexHash(newHash, hexHash(newHash, auth.Username+":"+state.Realm+":"+auth.Password)+":"+state.Nonce+":"+clientNonce)
	}
	credentials, err := digest.Digest(challenge, options)
	if err != nil {
		return "", errs.BadRequestError("the Digest challenge cannot be answered: "+err.Error(), "DIG-02400")
	}
	credentials.Algorithm = algorithm.String()
	header := credentials.String()

	logger.Debug(ctx, "Finished", map[string]any{"algorithm": algorithm, "nonceCount": state.NonceCount})
	return header, nil
}

func challengeAlgorithm(raw string) enums.DigestAlgorithm {
	if algorithm := digestalgorithm.GetDigestAlgorithm(raw); algorithm != digestalgorithm.Unknown {
		return algorithm
	}
	return digestalgorithm.MD5
}

func challengeQop(offered []string) enums.DigestQop {
	trimmed := make([]string, 0, len(offered))
	for _, qop := range offered {
		trimmed = append(trimmed, strings.TrimSpace(qop))
	}
	for _, preferred := range []enums.DigestQop{digestqop.Auth, digestqop.AuthInt} {
		for _, qop := range trimmed {
			if preferred.Equals(qop) {
				return preferred
			}
		}
	}
	return digestqop.None
}

func requestUri(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return parsed.RequestURI()
}

func digestHash(algorithm string) func() hash.Hash {
	switch strings.ToUpper(algorithm) {
	case "SHA-256":
		return sha256.New
	case "SHA-512-256":
		return sha512.New512_256
	default:
		return md5.New
	}
}

func hexHash(newHash func() hash.Hash, text string) string {
	sum := newHash()
	sum.Write([]byte(text))
	return hex.EncodeToString(sum.Sum(nil))
}

func newClientNonce() string {
	random := make([]byte, clientNonceLength)
	_, _ = rand.Read(random)
	return hex.EncodeToString(random)
}
