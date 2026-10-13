package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
)

func CollectScriptChain(owners []entity.ScriptOwner, phase enums.ScriptPhase) []entity.ScriptRef {
	start := 0
	for index, owner := range owners {
		if owner.CutsInheritance() {
			start = index
		}
	}
	chain := make([]entity.ScriptRef, 0, len(owners))
	for _, owner := range owners[start:] {
		if owner.Runs(phase) {
			chain = append(chain, owner.Ref(phase))
		}
	}
	return chain
}
