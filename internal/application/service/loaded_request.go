package service

import "github.com/brienze1/xenon/internal/domain/entity"

type loadedRequest struct {
	root    string
	request entity.Request
}
