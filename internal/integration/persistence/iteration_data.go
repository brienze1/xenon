package persistence

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	jsonDataExtension = ".json"
	csvDataExtension  = ".csv"
)

type iterationDataPersistence struct{}

func NewIterationDataPersistence() adapter.IterationDataPersistence {
	return &iterationDataPersistence{}
}

func (i *iterationDataPersistence) FindIterationData(ctx context.Context, path string) (*entity.IterationData, error) {
	logger.Debug(ctx, "Started", map[string]any{"path": path})

	content, err := os.ReadFile(path)
	if err != nil {
		logger.Warn(ctx, err, "The iteration data file cannot be read", map[string]any{"path": path})
		return nil, invalidIterationData("the file cannot be read", "RUN-05400")
	}
	dataModel, err := parseIterationData(path, content)
	if err != nil {
		return nil, err
	}
	data := dataModel.ToEntity(filepath.Base(path))

	logger.Debug(ctx, "Finished", map[string]any{"path": path, "bytes": len(content), "rows": data.Count()})
	return &data, nil
}

func (i *iterationDataPersistence) IterationDataExists(ctx context.Context, path string) bool {
	logger.Debug(ctx, "Started", map[string]any{"path": path})

	info, err := os.Stat(path)
	exists := err == nil && !info.IsDir()

	logger.Debug(ctx, "Finished", map[string]any{"path": path, "exists": exists})
	return exists
}

func parseIterationData(path string, content []byte) (*model.IterationData, error) {
	content = bytes.TrimPrefix(content, utf8ByteOrderMark)
	if !utf8.Valid(content) {
		return nil, invalidIterationData("the file is not UTF-8 text", "RUN-05400")
	}
	var rows []map[string]any
	var err error
	if isJSONData(path, content) {
		rows, err = parseJSONData(content)
	} else {
		rows, err = parseCSVData(content)
	}
	if err != nil {
		return nil, err
	}
	return model.NewIterationData(rows), nil
}

func isJSONData(path string, content []byte) bool {
	extension := strings.ToLower(filepath.Ext(path))
	if extension == jsonDataExtension {
		return true
	}
	trimmed := bytes.TrimSpace(content)
	return extension != csvDataExtension && len(trimmed) > 0 && (trimmed[0] == '[' || trimmed[0] == '{')
}

func parseJSONData(content []byte) ([]map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	var parsed any
	if err := decoder.Decode(&parsed); err != nil {
		return nil, invalidIterationData("the file is not valid JSON", "RUN-05400")
	}
	items, isArray := parsed.([]any)
	if !isArray {
		return nil, invalidIterationData("the JSON is not an array of objects", "RUN-05400")
	}
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		row, isObject := item.(map[string]any)
		if !isObject {
			return nil, invalidIterationData("the JSON is not an array of objects", "RUN-05400")
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseCSVData(content []byte) ([]map[string]any, error) {
	reader := csv.NewReader(bytes.NewReader(content))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, invalidIterationData("the CSV cannot be read", "RUN-05400")
	}
	if len(records) == 0 {
		return nil, invalidIterationData("the CSV has no header row", "RUN-05400")
	}
	header := records[0]
	rows := make([]map[string]any, 0, len(records)-1)
	for _, record := range records[1:] {
		row := make(map[string]any, len(header))
		for index, field := range header {
			row[field] = ""
			if index < len(record) {
				row[field] = record[index]
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func invalidIterationData(reason, code string) error {
	return errs.BadRequestError("the iteration data file cannot be used: "+reason, code,
		errs.ErrorDetails{Attribute: "reason", Messages: []string{reason}})
}
