package entity

import "strings"

const (
	defaultAwsRegion  = "us-east-1"
	defaultAwsService = "execute-api"
	awsS3Service      = "s3"
)

type AwsV4Auth struct {
	AccessKey          string
	SecretKey          string
	SessionToken       string
	Region             string
	Service            string
	AddAuthDataToQuery bool
}

func (a AwsV4Auth) RegionOrDefault() string {
	if strings.TrimSpace(a.Region) == "" {
		return defaultAwsRegion
	}
	return strings.TrimSpace(a.Region)
}

func (a AwsV4Auth) ServiceOrDefault() string {
	if strings.TrimSpace(a.Service) == "" {
		return defaultAwsService
	}
	return strings.TrimSpace(a.Service)
}

func (a AwsV4Auth) IsS3() bool {
	return strings.EqualFold(a.ServiceOrDefault(), awsS3Service)
}

func (a AwsV4Auth) IsBlank() bool {
	return a == AwsV4Auth{}
}

func (a AwsV4Auth) MapText(transform func(string) string) AwsV4Auth {
	a.AccessKey = transform(a.AccessKey)
	a.SecretKey = transform(a.SecretKey)
	a.SessionToken = transform(a.SessionToken)
	a.Region = transform(a.Region)
	a.Service = transform(a.Service)
	return a
}

func (a AwsV4Auth) TextFields(prefix string) []TextField {
	return []TextField{
		NewTextField(fieldName(prefix, "awsv4", "accessKey"), a.AccessKey),
		NewTextField(fieldName(prefix, "awsv4", "secretKey"), a.SecretKey),
		NewTextField(fieldName(prefix, "awsv4", "sessionToken"), a.SessionToken),
	}
}

func (a AwsV4Auth) LogValue() any {
	return map[string]any{"accessKey": a.AccessKey, "region": a.Region, "service": a.Service, "query": a.AddAuthDataToQuery, "hasSecretKey": a.SecretKey != "", "hasSessionToken": a.SessionToken != ""}
}
