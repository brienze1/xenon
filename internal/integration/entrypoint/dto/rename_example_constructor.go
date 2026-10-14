package dto

func NewRenameExample(localId, requestId, exampleId, name string) RenameExample {
	return RenameExample{LocalId: localId, RequestId: requestId, ExampleId: exampleId, Name: name}
}
