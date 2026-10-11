package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var sshAgentOnce sync.Once
var sshAgentInstance adapter.SshAgent

func GetSshAgent() adapter.SshAgent {
	if sshAgentInstance == nil {
		sshAgentOnce.Do(func() {
			sshAgentInstance = infraAdapter.NewSshAgent()
		})
	}
	return sshAgentInstance
}

func SetSshAgent(sshAgent adapter.SshAgent) {
	sshAgentOnce.Do(func() {})
	sshAgentInstance = sshAgent
}
