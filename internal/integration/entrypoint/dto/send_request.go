package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
)

type SendRequest struct {
	Origin                    SendOrigin `json:"-"`
	LocalId                   string     `json:"local_id" validate:"required"`
	TabId                     string     `json:"tab_id" validate:"required"`
	ConfirmedEnvironmentId    string     `json:"confirmed_environment_id"`
	Request                   Request    `json:"request"`
	AcknowledgeMissingSecrets bool       `json:"acknowledge_missing_secrets"`
	SendWithExpiredToken      bool       `json:"send_with_expired_token"`
}

type SendOrigin struct {
	Source string
	Client string
}

func UserSendOrigin() SendOrigin {
	return SendOrigin{Source: historysource.User.String()}
}

func AISendOrigin(client string) SendOrigin {
	return SendOrigin{Source: historysource.AI.String(), Client: client}
}

func (s SendOrigin) ToEntity() entity.SendOrigin {
	return entity.SendOrigin{Source: historysource.GetHistorySource(s.Source), Client: s.Client}.Normalized()
}

func (s SendRequest) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "tab_id": s.TabId, "request": s.Request.LogValue(), "acknowledge_missing_secrets": s.AcknowledgeMissingSecrets, "confirmed_environment_id": s.ConfirmedEnvironmentId, "origin": s.Origin.Source, "client": s.Origin.Client}
}

func (s *SendRequest) ToEntity() entity.Request {
	return s.Request.ToEntity()
}

func (s *SendRequest) ToAcknowledgement() entity.SendAcknowledgement {
	return entity.SendAcknowledgement{ConfirmedEnvironmentID: s.ConfirmedEnvironmentId, MissingSecrets: s.AcknowledgeMissingSecrets, ExpiredToken: s.SendWithExpiredToken}
}
