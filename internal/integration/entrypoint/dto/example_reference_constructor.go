package dto

func NewExampleReference(localId, requestId, exampleId string) ExampleReference {
	return ExampleReference{LocalId: localId, RequestId: requestId, ExampleId: exampleId}
}
