package entity

import "fmt"

type GitHubUser struct {
	Login     string
	Name      string
	Email     string
	AvatarURL string
	ID        int64
}

func (g GitHubUser) CommitName() string {
	if g.Name == "" {
		return g.Login
	}
	return g.Name
}

func (g GitHubUser) CommitEmail() string {
	if g.Email == "" {
		return fmt.Sprintf("%d+%s@users.noreply.github.com", g.ID, g.Login)
	}
	return g.Email
}
