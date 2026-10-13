package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type KnownHostsPersistence interface {
	usecase.FindKnownHosts
	usecase.TrustHostKey
}
