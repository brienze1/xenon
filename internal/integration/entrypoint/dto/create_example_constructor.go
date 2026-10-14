package dto

func NewCreateExample(localId, requestId, name string) CreateExample {
	return CreateExample{LocalId: localId, RequestId: requestId, Name: name}
}
