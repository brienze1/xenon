package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type AuthAwsV4 struct {
	AccessKey          string `yaml:"accessKey" json:"accessKey"`
	SecretKey          string `yaml:"secretKey" json:"secretKey"`
	SessionToken       string `yaml:"sessionToken" json:"sessionToken"`
	Region             string `yaml:"region" json:"region"`
	Service            string `yaml:"service" json:"service"`
	AddAuthDataToQuery bool   `yaml:"addAuthDataToQuery" json:"addAuthDataToQuery"`
}

func NewAuthAwsV4(auth entity.AwsV4Auth) AuthAwsV4 {
	return AuthAwsV4{AccessKey: auth.AccessKey, SecretKey: auth.SecretKey, SessionToken: auth.SessionToken, Region: auth.Region, Service: auth.Service, AddAuthDataToQuery: auth.AddAuthDataToQuery}
}

func (a *AuthAwsV4) ToEntity() entity.AwsV4Auth {
	return entity.AwsV4Auth{AccessKey: a.AccessKey, SecretKey: a.SecretKey, SessionToken: a.SessionToken, Region: a.Region, Service: a.Service, AddAuthDataToQuery: a.AddAuthDataToQuery}
}

func (a *AuthAwsV4) fields() []yamlnode.Field {
	return []yamlnode.Field{
		yamlnode.String("accessKey", a.AccessKey),
		yamlnode.String("secretKey", a.SecretKey),
		yamlnode.String("sessionToken", a.SessionToken),
		yamlnode.String("region", a.Region),
		yamlnode.String("service", a.Service),
		yamlnode.Bool("addAuthDataToQuery", a.AddAuthDataToQuery),
	}
}
