package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/gitprovider"
)

type AddGitToken struct {
	Provider    string `json:"provider" validate:"required,oneof=github gitlab bitbucket azure_devops other"`
	Host        string `json:"host" validate:"required,not_blank"`
	Username    string `json:"username"`
	Token       string `json:"token" validate:"required,not_blank"`
	Repository  string `json:"repository"`
	DisplayName string `json:"display_name"`
	CommitName  string `json:"commit_name"`
	CommitEmail string `json:"commit_email" validate:"omitempty,email"`
}

func (a *AddGitToken) ToEntity() entity.GitAccount {
	return entity.GitAccount{
		Provider: gitprovider.GetGitProvider(a.Provider), Host: a.Host, Username: a.Username,
		DisplayName: a.DisplayName, CommitName: a.CommitName, CommitEmail: a.CommitEmail,
	}
}

func (a AddGitToken) LogValue() any {
	return map[string]any{"provider": a.Provider, "host": a.Host, "username": a.Username, "repository": a.Repository, "hasToken": a.Token != ""}
}
