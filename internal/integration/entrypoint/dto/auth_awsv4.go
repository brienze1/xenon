package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type AuthAwsV4 struct {
	AccessKey          string `json:"access_key"`
	SecretKey          string `json:"secret_key"`
	SessionToken       string `json:"session_token"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	AddAuthDataToQuery bool   `json:"add_auth_data_to_query"`
}

func NewAuthAwsV4(auth entity.AwsV4Auth) *AuthAwsV4 {
	return &AuthAwsV4{AccessKey: auth.AccessKey, SecretKey: auth.SecretKey, SessionToken: auth.SessionToken, Region: auth.Region, Service: auth.Service, AddAuthDataToQuery: auth.AddAuthDataToQuery}
}

func (a *AuthAwsV4) ToEntity() entity.AwsV4Auth {
	if a == nil {
		return entity.AwsV4Auth{}
	}
	return entity.AwsV4Auth{AccessKey: a.AccessKey, SecretKey: a.SecretKey, SessionToken: a.SessionToken, Region: a.Region, Service: a.Service, AddAuthDataToQuery: a.AddAuthDataToQuery}
}
