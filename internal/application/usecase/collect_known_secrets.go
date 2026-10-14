package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func CollectKnownSecrets(vault *entity.Vault) []string {
	if vault == nil {
		return []string{}
	}
	return vault.KnownSecretValues()
}
