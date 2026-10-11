package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type GitRemoteInfo struct {
	DefaultBranch string   `json:"default_branch"`
	Branches      []string `json:"branches"`
}

func NewGitRemoteInfo(info entity.RemoteInfo) GitRemoteInfo {
	return GitRemoteInfo{DefaultBranch: info.DefaultBranch, Branches: info.Branches}
}
