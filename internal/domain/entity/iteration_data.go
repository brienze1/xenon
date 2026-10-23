package entity

type IterationData struct {
	FileName string
	Rows     []map[string]any
}

func (i IterationData) Count() int {
	return len(i.Rows)
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
