package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type AwsV4Webservice interface {
	usecase.SignAwsRequest
}
