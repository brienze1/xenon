package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type PostmanWebservice interface {
	usecase.ReadPostmanCollection
	usecase.ReadPostmanEnvironment
	usecase.ReadPostmanDump
}
