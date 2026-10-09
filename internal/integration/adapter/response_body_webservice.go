package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type ResponseBodyWebservice interface {
	usecase.SaveResponse
	usecase.FindResponse
	usecase.ReleaseResponse
	usecase.ExportResponseBody
}
