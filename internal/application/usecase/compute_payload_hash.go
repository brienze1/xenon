package usecase

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/brienze1/xenon/internal/domain/entity"
)

// ComputePayloadHash hashes the body as AWS Signature v4 signs it; form-data bodies are signed as empty, as in Postman.
func ComputePayloadHash(body entity.Body, content []byte) string {
	if body.IsMultipart() {
		content = nil
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
