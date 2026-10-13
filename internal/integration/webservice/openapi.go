package webservice

import (
	"context"
	"errors"
	"regexp"
	"strconv"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/openapi/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

var yamlErrorLine = regexp.MustCompile(`line (\d+): (.*)`)

type openApiWebservice struct{}

func NewOpenApiWebservice() adapter.OpenApiWebservice {
	return &openApiWebservice{}
}

func (o *openApiWebservice) WriteDocument(ctx context.Context, document entity.ApiDocument) ([]byte, error) {
	logger.Debug(ctx, "Started", map[string]any{"title": document.Title, "operations": len(document.Operations)})

	data, err := dto.EncodeDocument(dto.NewDocument(document))
	if err != nil {
		return nil, errs.InternalServerError(err, "DOC-08500", "the OpenAPI document could not be written")
	}

	logger.Debug(ctx, "Finished", map[string]any{"title": document.Title, "bytes": len(data)})
	return data, nil
}

func (o *openApiWebservice) ReadDocument(ctx context.Context, data []byte) (*entity.ApiDocument, error) {
	logger.Debug(ctx, "Started", map[string]any{"bytes": len(data)})

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, yamlProblem(data, err, "DOC-01400")
	}
	if err := validateOpenApi(&root); err != nil {
		return nil, err
	}
	var document dto.Document
	if err := root.Decode(&document); err != nil {
		return nil, yamlProblem(data, err, "DOC-09400")
	}
	read, err := document.ToEntity()
	if err != nil {
		return nil, yamlProblem(data, err, "DOC-10400")
	}

	logger.Debug(ctx, "Finished", map[string]any{"title": read.Title, "operations": len(read.Operations)})
	return &read, nil
}

func yamlProblem(data []byte, err error, code string) error {
	message := err.Error()
	var typeError *yaml.TypeError
	if errors.As(err, &typeError) && len(typeError.Errors) > 0 {
		message = typeError.Errors[0]
	}
	line := 1
	if match := yamlErrorLine.FindStringSubmatch(message); match != nil {
		line, _ = strconv.Atoi(match[1])
		message = match[2]
	}
	return documentProblem(code, message, line, firstColumn(data, line))
}

func documentProblem(code, message string, line, column int) error {
	return errs.BadRequestError("line "+strconv.Itoa(line)+", column "+strconv.Itoa(column)+": "+message, code,
		errs.ErrorDetails{Attribute: "line", Messages: []string{strconv.Itoa(line)}},
		errs.ErrorDetails{Attribute: "column", Messages: []string{strconv.Itoa(column)}},
	)
}

func firstColumn(data []byte, line int) int {
	current, column := 1, 1
	for _, char := range data {
		if current > line {
			break
		}
		if char == '\n' {
			current, column = current+1, 1
			continue
		}
		if current == line && char != ' ' && char != '\t' {
			return column
		}
		column++
	}
	return 1
}
