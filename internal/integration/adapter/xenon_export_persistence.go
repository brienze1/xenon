package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type XenonExportPersistence interface {
	usecase.WriteXenonExport
	usecase.EncryptExportSecrets
	usecase.ReadXenonExport
	usecase.DecryptExportSecrets
}
