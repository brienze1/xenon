package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
)

const DefaultAIClientName = "AI client"

type SendOrigin struct {
	Source  enums.HistorySource
	Client  string
	Session string
}

func UserSendOrigin() SendOrigin {
	return SendOrigin{Source: historysource.User}
}

func AISendOrigin(client, session string) SendOrigin {
	return SendOrigin{Source: historysource.AI, Client: AIClientName(client), Session: session}
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
		return AISendOrigin(s.Client, s.Session)
	}
	return UserSendOrigin()
}

func (s SendOrigin) LogValue() any {
	return map[string]any{"source": s.Source, "client": s.Client, "session": s.Session}
}
