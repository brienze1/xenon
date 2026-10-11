package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/signinstatus"
)

type GitHubSignIn struct {
	Status   enums.SignInStatus
	Token    string
	Interval int
}

func (g GitHubSignIn) IsDone() bool {
	return g.Status == signinstatus.Done
}

func (g GitHubSignIn) LogValue() any {
	return map[string]any{"status": g.Status, "interval": g.Interval, "hasToken": g.Token != ""}
}
