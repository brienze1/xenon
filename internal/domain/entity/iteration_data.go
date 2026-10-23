package entity

import "github.com/brienze1/xenon/pkg/errs"

const noIterationDataRows = "the file has no rows"

type IterationData struct {
	FileName string
	Rows     []map[string]any
}

func (i IterationData) Validate() error {
	if len(i.Rows) == 0 {
		return errs.BadRequestError("the iteration data file cannot be used: "+noIterationDataRows, "RUN-05400",
			errs.ErrorDetails{Attribute: "reason", Messages: []string{noIterationDataRows}})
	}
	return nil
}

func (i IterationData) Count() int {
	return len(i.Rows)
}

// IterationCount is one iteration per data row, and one when there is no data.
func (i IterationData) IterationCount() int {
	return max(1, i.Count())
}

// Row is the data row of an iteration; iterations past the last row reuse it.
func (i IterationData) Row(iteration int) map[string]any {
	if len(i.Rows) == 0 {
		return nil
	}
	return i.Rows[min(iteration, len(i.Rows)-1)]
}

func (i IterationData) WrappedRow(n int) map[string]any {
	if len(i.Rows) == 0 {
		return nil
	}
	return i.Rows[n%len(i.Rows)]
}

// IterationDataText is a data value as {{field}} sends it: null is empty, objects and arrays are compact JSON.
func IterationDataText(value any) string {
	if value == nil {
		return ""
	}
	return LocalVariableText(value)
}
