package entity

// DocumentedRequest is a request of a collection with the folders above it and the auth it really uses.
type DocumentedRequest struct {
	Folders []Collection
	Auth    Auth
	Request Request
}

// NewDocumentedRequest takes the ancestors outermost first, the collection itself included.
func NewDocumentedRequest(request Request, ancestors []Collection) DocumentedRequest {
	return DocumentedRequest{Request: request, Auth: effectiveAuth(request.Auth, ancestors), Folders: foldersOf(ancestors)}
}

func (d DocumentedRequest) Tag() string {
	return NewApiTag(d.Folders).Name
}

func effectiveAuth(auth Auth, ancestors []Collection) Auth {
	for index := len(ancestors) - 1; index >= 0 && auth.IsInherited(); index-- {
		auth = ancestors[index].Auth
	}
	return auth
}

func foldersOf(ancestors []Collection) []Collection {
	if len(ancestors) < 2 {
		return []Collection{}
	}
	return append([]Collection{}, ancestors[1:]...)
}
