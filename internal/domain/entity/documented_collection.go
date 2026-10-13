package entity

import "slices"

const documentVersion = "1.0.0"

// DocumentedCollection gathers a collection's or folder's sub-folders and requests, in tree order, to document them.
// Collection is the documented collection or folder; Inherited holds the containers above it, outermost first.
type DocumentedCollection struct {
	Tags       []ApiTag
	TagGroups  []ApiTagGroup
	Requests   []DocumentedRequest
	Inherited  []Collection
	Collection Collection
}

func NewDocumentedCollection(collection Collection, inherited []Collection) DocumentedCollection {
	return DocumentedCollection{
		Collection: collection,
		Inherited:  slices.Clone(inherited),
		Tags:       []ApiTag{},
		TagGroups:  []ApiTagGroup{},
		Requests:   []DocumentedRequest{},
	}
}

// WithFolder takes the ancestors from the documented collection or folder down to the folder itself.
func (d DocumentedCollection) WithFolder(ancestors []Collection) DocumentedCollection {
	tag := NewApiTag(foldersOf(ancestors))
	d.Tags = append(slices.Clone(d.Tags), tag)
	d.TagGroups = WithTagGroup(d.TagGroups, foldersOf(ancestors), tag)
	return d
}

// WithRequest takes the ancestors from the documented collection or folder down to the request's folder.
func (d DocumentedCollection) WithRequest(request Request, ancestors []Collection) DocumentedCollection {
	d.Requests = append(slices.Clone(d.Requests), NewDocumentedRequest(request, d.Inherited, ancestors))
	return d
}

func (d DocumentedCollection) Request(id string) (DocumentedRequest, bool) {
	index := slices.IndexFunc(d.Requests, func(documented DocumentedRequest) bool { return documented.Request.ID == id })
	if index < 0 {
		return DocumentedRequest{}, false
	}
	return d.Requests[index], true
}

// ApiDocument builds the OpenAPI view of the collection or folder: one operation per request, grouped by path.
func (d DocumentedCollection) ApiDocument() ApiDocument {
	document := ApiDocument{
		Title:           d.Collection.Name,
		Description:     d.Collection.Description,
		Version:         documentVersion,
		Servers:         []string{},
		Tags:            d.Tags,
		TagGroups:       d.TagGroups,
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
