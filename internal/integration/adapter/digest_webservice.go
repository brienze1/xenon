package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type DigestWebservice interface {
	usecase.AnswerDigestChallenge
}
