package controller

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/application"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/controller"
)

var vaultControllerOnce sync.Once
var vaultControllerInstance adapter.VaultController

func GetVaultController() adapter.VaultController {
	if vaultControllerInstance == nil {
		vaultControllerOnce.Do(func() {
			vaultControllerInstance = controller.NewVault(GetErrorHandler(), application.GetManageVaultService())
		})
	}
	return vaultControllerInstance
}
