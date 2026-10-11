package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type OAuth2Webservice interface {
	usecase.ExchangeOAuthToken
	usecase.RefreshOAuthToken
}
