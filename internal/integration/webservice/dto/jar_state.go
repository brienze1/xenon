package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/webservice/cookiejar"
)

type JarState struct {
	Entries       []cookiejar.Entry `json:"entries"`
	Revision      uint64            `json:"revision"`
	SavedRevision uint64            `json:"saved_revision"`
	Loaded        bool              `json:"loaded"`
}

func NewJarState(entries []cookiejar.Entry, revision, savedRevision uint64, loaded bool) JarState {
	return JarState{Entries: entries, Revision: revision, SavedRevision: savedRevision, Loaded: loaded}
}

func (j JarState) ToEntity() entity.CookieJarState {
	return entity.CookieJarState{
		Cookies:       NewJarCookies(j.Entries),
		Revision:      j.Revision,
		SavedRevision: j.SavedRevision,
		Loaded:        j.Loaded,
	}
}
