package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
)

const DefaultAIClientName = "AI client"

type SendOrigin struct {
	Source enums.HistorySource
	Client string
}

func UserSendOrigin() SendOrigin {
	return SendOrigin{Source: historysource.User}
}

func AISendOrigin(client string) SendOrigin {
	return SendOrigin{Source: historysource.AI, Client: AIClientName(client)}
}

func AIClientName(client string) string {
	if trimmed := strings.TrimSpace(client); trimmed != "" {
		return trimmed
	}
	return DefaultAIClientName
}

func (s SendOrigin) IsAI() bool {
	return s.Source == historysource.AI
}

func (s SendOrigin) Normalized() SendOrigin {
	if s.IsAI() {
		return AISendOrigin(s.Client)
	}
	return UserSendOrigin()
}

func (s SendOrigin) LogValue() any {
	return map[string]any{"source": s.Source, "client": s.Client}
}
