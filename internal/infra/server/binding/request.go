package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type RequestService struct {
	get      adapter.RequestGetController
	save     adapter.RequestSaveController
	edit     adapter.RequestEditController
	send     adapter.RequestSendController
	response adapter.RequestResponseController
}

func NewRequestService() *RequestService {
	return &RequestService{
		get:      controller.GetRequestGet(),
		save:     controller.GetRequestSave(),
		edit:     controller.GetRequestEdit(),
		send:     controller.GetRequestSend(),
		response: controller.GetRequestResponse(),
	}
}

func (r *RequestService) Get(ctx context.Context, request dto.OpenRequest) (*dto.Request, error) {
	return r.get.Get(ctx, request)
}

func (r *RequestService) Save(ctx context.Context, request dto.SaveRequest) (*dto.Request, error) {
	return r.save.Save(ctx, request)
}

func (r *RequestService) SaveAs(ctx context.Context, request dto.SaveRequestAs) (*dto.Request, error) {
	return r.save.SaveAs(ctx, request)
}

func (r *RequestService) ApplyUrl(ctx context.Context, request dto.ApplyRequest) (*dto.Request, error) {
	return r.edit.ApplyUrl(ctx, request)
}

func (r *RequestService) ApplyParams(ctx context.Context, request dto.ApplyRequest) (*dto.Request, error) {
	return r.edit.ApplyParams(ctx, request)
}

func (r *RequestService) ParseBulk(ctx context.Context, request dto.BulkText) (*dto.BulkRows, error) {
	return r.edit.ParseBulk(ctx, request)
}

func (r *RequestService) FormatBulk(ctx context.Context, request dto.BulkRows) (*dto.BulkText, error) {
	return r.edit.FormatBulk(ctx, request)
}

func (r *RequestService) AutoHeaders(ctx context.Context, request dto.AutoHeadersRequest) (dto.AutoHeaderList, error) {
	return r.edit.AutoHeaders(ctx, request)
}

func (r *RequestService) Send(ctx context.Context, request dto.SendRequest) (*dto.SendResult, error) {
	return r.send.Send(ctx, request)
}

func (r *RequestService) LastResult(ctx context.Context, request dto.TabReference) (*dto.SendResult, error) {
	return r.response.LastResult(ctx, request)
}

func (r *RequestService) ReleaseResponse(ctx context.Context, request dto.TabReference) error {
	return r.response.ReleaseResponse(ctx, request)
}

func (r *RequestService) SaveResponseBody(ctx context.Context, request dto.SaveResponseBody) error {
	return r.response.SaveResponseBody(ctx, request)
}

func (r *RequestService) ServiceShutdown() error {
	return r.response.ReleaseAll(context.Background())
}
