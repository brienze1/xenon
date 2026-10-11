package sandbox

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"

	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	referenceFile = "reference/reference.json"
	snippetsFile  = "reference/snippets.json"
)

//go:embed reference/reference.json reference/snippets.json
var referenceFiles embed.FS

type scriptToolsWebservice struct {
	files fs.FS
}

func NewScriptToolsWebservice() adapter.ScriptToolsWebservice {
	return &scriptToolsWebservice{files: referenceFiles}
}

func (s *scriptToolsWebservice) Check(ctx context.Context, text string) []entity.ScriptError {
	logger.Debug(ctx, "Started", map[string]any{"length": len(text)})

	errors := []entity.ScriptError{}
	if _, err := goja.Compile(scriptFile, wrap(text), false); err != nil {
		errors = append(errors, syntaxError(err, entity.ScriptRef{Text: text}))
	}

	logger.Debug(ctx, "Finished", map[string]any{"length": len(text), "errors": len(errors)})
	return errors
}

func (s *scriptToolsWebservice) FindReference(ctx context.Context) ([]entity.ScriptMember, error) {
	logger.Debug(ctx, "Started")

	var reference dto.ScriptReference
	if err := s.read(referenceFile, &reference, "SCR-06500", "SCR-07500"); err != nil {
		return nil, err
	}
	members := reference.ToEntity()

	logger.Debug(ctx, "Finished", map[string]any{"members": len(members)})
	return members, nil
}

func (s *scriptToolsWebservice) FindSnippets(ctx context.Context) ([]entity.ScriptSnippet, error) {
	logger.Debug(ctx, "Started")

	var stored []dto.ScriptSnippet
	if err := s.read(snippetsFile, &stored, "SCR-08500", "SCR-09500"); err != nil {
		return nil, err
	}
	snippets := make([]entity.ScriptSnippet, 0, len(stored))
	for _, snippet := range stored {
		snippets = append(snippets, snippet.ToEntity())
	}

	logger.Debug(ctx, "Finished", map[string]any{"snippets": len(snippets)})
	return snippets, nil
}

func (s *scriptToolsWebservice) read(name string, target any, readCode, parseCode string) error {
	data, err := fs.ReadFile(s.files, name)
	if err != nil {
		return errs.InternalServerError(err, readCode, "the script editor data could not be read")
	}
	if err := json.Unmarshal(data, target); err != nil {
		return errs.InternalServerError(err, parseCode, "the script editor data is not valid")
	}
	return nil
}
