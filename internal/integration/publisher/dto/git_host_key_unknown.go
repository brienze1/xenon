package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const GitHostKeyUnknownEvent = "git:host-key-unknown"

type GitHostKeyUnknown struct {
	Repository  string `json:"repository"`
	Host        string `json:"host"`
	Fingerprint string `json:"fingerprint"`
	KeyType     string `json:"key_type"`
}

func NewGitHostKeyUnknown(repository string, key entity.HostKey) GitHostKeyUnknown {
	return GitHostKeyUnknown{Repository: repository, Host: key.Host, Fingerprint: key.Fingerprint, KeyType: key.KeyType}
}
