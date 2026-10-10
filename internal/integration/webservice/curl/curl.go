package curl

import (
	"context"
	"regexp"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

var urlScheme = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9+.\-]*)://`)

type curlWebservice struct{}

func NewCurlWebservice() adapter.CurlWebservice {
	return &curlWebservice{}
}

func (c *curlWebservice) ReadCommand(ctx context.Context, text string) (*entity.ParsedCurl, error) {
	logger.Debug(ctx, "Started", map[string]any{"length": len(text)})

	arguments, err := tokenize(normalize(strings.TrimSpace(text)))
	if err != nil {
		return nil, err
	}
	command := parseOptions(arguments)
	address, err := httpUrl(command.Urls)
	if err != nil {
		return nil, err
	}
	for _, other := range command.Urls {
		if other != address {
			command.Notice("Ignored the extra URL " + other)
		}
	}
	parsed := command.ToEntity(address)

	logger.Debug(ctx, "Finished", map[string]any{"length": len(text), "result": parsed})
	return &parsed, nil
}

func httpUrl(urls []string) (string, error) {
	if len(urls) == 0 {
		return "", errs.BadRequestError("The cURL command has no URL", "IMP-06400")
	}
	for _, address := range urls {
		if isHttp(address) {
			return address, nil
		}
	}
	return "", errs.BadRequestError("Only HTTP and HTTPS URLs can be imported, not "+urls[0], "IMP-07400")
}

func isHttp(address string) bool {
	match := urlScheme.FindStringSubmatch(address)
	if match == nil {
		return true
	}
	scheme := strings.ToLower(match[1])
	return scheme == "http" || scheme == "https"
}
