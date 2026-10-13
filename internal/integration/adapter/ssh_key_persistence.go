package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type SshKeyPersistence interface {
	usecase.GenerateSshKey
	usecase.ReadSshKey
	usecase.DeleteSshKey
}
