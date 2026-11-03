package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
)

type SendOrigin struct {
	Source  string `json:"source"`
	Client  string `json:"client"`
	Session string `json:"session"`
}

func NewSendOrigin(source enums.HistorySource, client, session string) SendOrigin {
	return SendOrigin{Source: source.String(), Client: client, Session: session}
}

func (s SendOrigin) ToEntity() entity.SendOrigin {
	return entity.SendOrigin{Source: historysource.GetHistorySource(s.Source), Client: s.Client, Session: s.Session}.Normalized()
}
