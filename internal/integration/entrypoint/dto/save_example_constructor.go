package dto

func NewSaveExample(localId, requestId, baseRevision string, example ExampleEdit) SaveExample {
	return SaveExample{LocalId: localId, RequestId: requestId, BaseRevision: baseRevision, Example: example}
}
