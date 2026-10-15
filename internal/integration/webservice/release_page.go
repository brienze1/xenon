package webservice

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type releasePageWebservice struct {
	browser adapter.Browser
}

func NewReleasePageWebservice(browser adapter.Browser) adapter.ReleasePageWebservice {
	return &releasePageWebservice{browser: browser}
}

func (r *releasePageWebservice) OpenReleasePage(ctx context.Context, url string) error {
	logger.Debug(ctx, "Started", map[string]any{"url": url})

	if err := r.browser.OpenURL(url); err != nil {
		return errs.ServiceUnavailableError(err, "UPD-06503", "The release page could not be opened")
	}

	logger.Debug(ctx, "Finished", map[string]any{"url": url})
	return nil
}
