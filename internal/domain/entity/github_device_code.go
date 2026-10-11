package entity

type GitHubDeviceCode struct {
	SessionID       string
	UserCode        string
	VerificationURI string
	ExpiresIn       int
	Interval        int
}
