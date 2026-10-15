package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SendRequest struct {
	Origin                    SendOrigin `json:"-"`
	LocalId                   string     `json:"local_id" validate:"required"`
	TabId                     string     `json:"tab_id" validate:"required"`
	AliasId                   string     `json:"alias_id"`
	ConfirmedEnvironmentId    string     `json:"confirmed_environment_id"`
	Request                   Request    `json:"request"`
	AcknowledgeMissingSecrets bool       `json:"acknowledge_missing_secrets"`
	SendWithExpiredToken      bool       `json:"send_with_expired_token"`
}

func NewSendRequest(origin SendOrigin, localId, tabId, confirmedEnvironmentId string, request Request, acknowledgeMissingSecrets, sendWithExpiredToken bool) SendRequest {
	return SendRequest{
		Origin:                    origin,
		LocalId:                   localId,
		TabId:                     tabId,
		ConfirmedEnvironmentId:    confirmedEnvironmentId,
		Request:                   request,
		AcknowledgeMissingSecrets: acknowledgeMissingSecrets,
		SendWithExpiredToken:      sendWithExpiredToken,
	}
}

func (s SendRequest) LogValue() any {
	return map[string]any{"local_id": s.LocalId, "tab_id": s.TabId, "alias_id": s.AliasId, "request": s.Request.LogValue(), "acknowledge_missing_secrets": s.AcknowledgeMissingSecrets, "confirmed_environment_id": s.ConfirmedEnvironmentId, "origin": s.Origin.ToEntity().Source.String(), "client": s.Origin.Client}
}

func (s SendRequest) WithAlias(aliasId string) SendRequest {
	s.AliasId = aliasId
	return s
}

func (s *SendRequest) ToEntity() entity.Request {
	return s.Request.ToEntity()
}

func (s *SendRequest) ToAcknowledgement() entity.SendAcknowledgement {
	return entity.SendAcknowledgement{ConfirmedEnvironmentID: s.ConfirmedEnvironmentId, MissingSecrets: s.AcknowledgeMissingSecrets, ExpiredToken: s.SendWithExpiredToken}
}

func (s *SendRequest) ToSendOrigin() entity.SendOrigin {
	return s.Origin.ToEntity()
}

func (s *SendRequest) WorkspaceId() string {
	return s.LocalId
}

func (s *SendRequest) TargetTabId() string {
	return s.TabId
}
