package digestalgorithm

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	MD5           enums.DigestAlgorithm = "MD5"
	MD5Sess       enums.DigestAlgorithm = "MD5-sess"
	SHA256        enums.DigestAlgorithm = "SHA-256"
	SHA256Sess    enums.DigestAlgorithm = "SHA-256-sess"
	SHA512256     enums.DigestAlgorithm = "SHA-512-256"
	SHA512256Sess enums.DigestAlgorithm = "SHA-512-256-sess"
	Unknown       enums.DigestAlgorithm = "unknown"
)

func GetDigestAlgorithm(s string) enums.DigestAlgorithm {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "md5":
		return MD5
	case "md5-sess":
		return MD5Sess
	case "sha-256":
		return SHA256
	case "sha-256-sess":
		return SHA256Sess
	case "sha-512-256":
		return SHA512256
	case "sha-512-256-sess":
		return SHA512256Sess
	default:
		return Unknown
	}
}
