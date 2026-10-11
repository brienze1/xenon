package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type GitHubUserResponse struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	ID        int64  `json:"id"`
}

func (g *GitHubUserResponse) ToEntity(emails []GitHubEmail) entity.GitHubUser {
	user := entity.GitHubUser{ID: g.ID, Login: g.Login, Name: g.Name, AvatarURL: g.AvatarURL}
	for _, email := range emails {
		if email.Primary && email.Verified {
			user.Email = email.Email
		}
	}
	return user
}
