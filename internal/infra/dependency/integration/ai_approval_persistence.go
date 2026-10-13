package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var aiApprovalPersistenceOnce sync.Once
var aiApprovalPersistenceInstance adapter.AIApprovalPersistence

func GetAIApprovalPersistence() adapter.AIApprovalPersistence {
	if aiApprovalPersistenceInstance == nil {
		aiApprovalPersistenceOnce.Do(func() {
			aiApprovalPersistenceInstance = persistence.NewAIApprovalPersistence()
		})
	}
	return aiApprovalPersistenceInstance
}
