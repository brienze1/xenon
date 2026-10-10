package binding

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/server/binding"
)

var vaultServiceOnce sync.Once
var vaultServiceInstance *binding.VaultService

func GetVaultService() *binding.VaultService {
	if vaultServiceInstance == nil {
		vaultServiceOnce.Do(func() {
			vaultServiceInstance = binding.NewVaultService(controller.GetVaultController())
		})
	}
	return vaultServiceInstance
}
