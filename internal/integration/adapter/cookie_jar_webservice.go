package adapter

import (
	"net/http"

	"github.com/brienze1/xenon/internal/application/usecase"
)

type CookieJarWebservice interface {
	usecase.ManageCookieJar
	// HttpJar is the jar an http.Client of the workspace uses; without sendCookies it only stores.
	HttpJar(localId string, sendCookies bool) http.CookieJar
}
