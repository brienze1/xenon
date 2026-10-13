package gitprovider

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
)

const (
	GitHub      enums.GitProvider = "github"
	GitLab      enums.GitProvider = "gitlab"
	Bitbucket   enums.GitProvider = "bitbucket"
	AzureDevOps enums.GitProvider = "azure_devops"
	Other       enums.GitProvider = "other"
	Unknown     enums.GitProvider = "unknown"
)

func GetGitProvider(s string) enums.GitProvider {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "github":
		return GitHub
	case "gitlab":
		return GitLab
	case "bitbucket":
		return Bitbucket
	case "azure_devops":
		return AzureDevOps
	case "other":
		return Other
	default:
		return Unknown
	}
}
