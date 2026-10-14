package entity

import "net/url"

// RemoteWithoutCredentials drops any user info embedded in a repository URL; anything unparsable is returned as given.
func RemoteWithoutCredentials(repository string) string {
	parsed, err := url.Parse(repository)
	if err != nil || parsed.User == nil {
		return repository
	}
	parsed.User = nil
	return parsed.String()
}
