package entity

import (
	"encoding/json"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

func TestIterationDataRowReusesTheLastRowAndWrappedRowWraps(t *testing.T) {
	data := IterationData{Rows: []map[string]any{{"n": "a"}, {"n": "b"}}}
	if data.Count() != 2 || data.Row(0)["n"] != "a" || data.Row(5)["n"] != "b" {
		t.Errorf("rows = %v, %v", data.Row(0), data.Row(5))
	}
	if data.WrappedRow(2)["n"] != "a" || data.WrappedRow(3)["n"] != "b" {
		t.Errorf("wrapped = %v, %v", data.WrappedRow(2), data.WrappedRow(3))
	}
	if (IterationData{}).Row(0) != nil || (IterationData{}).WrappedRow(1) != nil {
		t.Error("no data has no rows")
	}
}

func TestIterationDataTextIsTheValueAsWritten(t *testing.T) {
	cases := map[string]any{"": nil, "text": "text", "1.5": json.Number("1.5"), "true": true, `{"a":1}`: map[string]any{"a": json.Number("1")}, "[1,2]": []any{json.Number("1"), json.Number("2")}}
	for expected, value := range cases {
		if got := IterationDataText(value); got != expected {
			t.Errorf("IterationDataText(%v) = %q, expected %q", value, got, expected)
		}
	}
}

func TestIterationDataValidateRefusesAFileWithoutRows(t *testing.T) {
	err := IterationData{FileName: "header.csv", Rows: []map[string]any{}}.Validate()
	failure, isBase := err.(errs.BaseError)
	if !isBase || failure.Code() != "RUN-05400" || failure.Description() != "the iteration data file cannot be used: the file has no rows" {
		t.Errorf("err = %v", err)
	}
	if err := (IterationData{Rows: []map[string]any{{"email": "a"}}}).Validate(); err != nil {
		t.Errorf("a file with rows is usable: %v", err)
	}
}
