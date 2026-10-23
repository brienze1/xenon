package usecase

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const jsonDataExtension = ".json"

var utf8ByteOrderMark = []byte{0xEF, 0xBB, 0xBF}

// ParseIterationData reads a CSV file with a header row, or a JSON array of objects, as UTF-8 with an optional byte-order mark.
func ParseIterationData(path string, content []byte) (*entity.IterationData, error) {
	content = bytes.TrimPrefix(content, utf8ByteOrderMark)
	if !utf8.Valid(content) {
		return nil, RejectInvalidIterationData("the file is not UTF-8 text")
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
	return &entity.IterationData{FileName: filepath.Base(path), Rows: rows}, nil
}

func isJSONData(path string, content []byte) bool {
	extension := strings.ToLower(filepath.Ext(path))
	if extension == jsonDataExtension {
		return true
	}
	trimmed := bytes.TrimSpace(content)
	return extension != ".csv" && len(trimmed) > 0 && (trimmed[0] == '[' || trimmed[0] == '{')
}

func parseJSONData(content []byte) ([]map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	var parsed any
	if err := decoder.Decode(&parsed); err != nil {
		return nil, RejectInvalidIterationData("the file is not valid JSON")
	}
	items, isArray := parsed.([]any)
	if !isArray {
		return nil, RejectInvalidIterationData("the JSON is not an array of objects")
	}
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		row, isObject := item.(map[string]any)
		if !isObject {
			return nil, RejectInvalidIterationData("the JSON is not an array of objects")
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, RejectInvalidIterationData("the file has no rows")
	}
	return rows, nil
}

func parseCSVData(content []byte) ([]map[string]any, error) {
	reader := csv.NewReader(bytes.NewReader(content))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, RejectInvalidIterationData("the CSV cannot be read")
	}
	if len(records) == 0 {
		return nil, RejectInvalidIterationData("the CSV has no header row")
	}
	if len(records) == 1 {
		return nil, RejectInvalidIterationData("the file has no rows")
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
