package entity

import "slices"

const documentVersion = "1.0.0"

// DocumentedCollection gathers a collection's folders and requests, in tree order, to document them.
type DocumentedCollection struct {
	Tags       []ApiTag
	Requests   []DocumentedRequest
	Collection Collection
}

func NewDocumentedCollection(collection Collection) DocumentedCollection {
	return DocumentedCollection{Collection: collection, Tags: []ApiTag{}, Requests: []DocumentedRequest{}}
}

func (d DocumentedCollection) WithFolder(ancestors []Collection) DocumentedCollection {
	d.Tags = append(slices.Clone(d.Tags), NewApiTag(foldersOf(ancestors)))
	return d
}

func (d DocumentedCollection) WithRequest(request Request, ancestors []Collection) DocumentedCollection {
	d.Requests = append(slices.Clone(d.Requests), NewDocumentedRequest(request, ancestors))
	return d
}

func (d DocumentedCollection) Request(id string) (DocumentedRequest, bool) {
	index := slices.IndexFunc(d.Requests, func(documented DocumentedRequest) bool { return documented.Request.ID == id })
	if index < 0 {
		return DocumentedRequest{}, false
	}
	return d.Requests[index], true
}

// ApiDocument builds the OpenAPI view of the collection: one operation per request, grouped by path.
func (d DocumentedCollection) ApiDocument() ApiDocument {
	document := ApiDocument{
		Title:           d.Collection.Name,
		Description:     d.Collection.Description,
		Version:         documentVersion,
		Servers:         []string{},
		Tags:            d.Tags,
		Operations:      []ApiOperation{},
		SecuritySchemes: []ApiSecurityScheme{},
	}
	origins, paths := d.splitUrls()
	prefix := ""
	for _, origin := range origins {
		if !slices.Contains(document.Servers, origin) && origin != "" {
			document.Servers = append(document.Servers, origin)
		}
	}
	if len(document.Servers) == 1 && !slices.Contains(origins, "") {
		prefix = commonPathPrefix(paths)
		document.Servers[0] += prefix
	}
	operations := make([]ApiOperation, 0, len(d.Requests))
	for index, documented := range d.Requests {
		var security []string
		document.SecuritySchemes, security = withSecurityScheme(document.SecuritySchemes, documented.Auth)
		operation := NewApiOperation(documented, trimPathPrefix(paths[index], prefix), security)
		if len(document.Servers) > 1 && origins[index] != document.Servers[0] {
			operation.Server = origins[index]
		}
		operations = append(operations, operation)
	}
	document.Operations = inDocumentOrder(operations)
	return document
}

func (d DocumentedCollection) splitUrls() ([]string, []string) {
	origins := make([]string, 0, len(d.Requests))
	paths := make([]string, 0, len(d.Requests))
	for _, documented := range d.Requests {
		origin, path := SplitRequestUrl(documented.Request.URL)
		origins = append(origins, origin)
		paths = append(paths, path)
	}
	return origins, paths
}

func withSecurityScheme(schemes []ApiSecurityScheme, auth Auth) ([]ApiSecurityScheme, []string) {
	scheme, secured := NewApiSecurityScheme(auth)
	if !secured {
		return schemes, []string{}
	}
	if index := slices.IndexFunc(schemes, scheme.Equal); index >= 0 {
		return schemes, []string{schemes[index].Name}
	}
	scheme = scheme.named(schemes)
	return append(schemes, scheme), []string{scheme.Name}
}

func inDocumentOrder(operations []ApiOperation) []ApiOperation {
	paths := make([]string, 0)
	for _, operation := range operations {
		if !slices.Contains(paths, operation.Path) {
			paths = append(paths, operation.Path)
		}
	}
	ordered := make([]ApiOperation, 0, len(operations))
	for _, path := range paths {
		methods := make([]string, 0)
		for _, operation := range operations {
			if operation.Path == path && !slices.Contains(methods, operation.Method) {
				methods = append(methods, operation.Method)
			}
		}
		for _, method := range methods {
			first := true
			for _, operation := range operations {
				if operation.Path == path && operation.Method == method {
					operation.Variant = !first
					ordered = append(ordered, operation)
					first = false
				}
			}
		}
	}
	return ordered
}
