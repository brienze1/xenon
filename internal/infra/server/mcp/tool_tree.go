package mcp

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

const (
	treePathSeparator   = " / "
	treeAllDepth        = "all"
	treeUnlimitedDepth  = -1
	treeSearchLimit     = 50
	treeCollectionScope = "collection:"
	treeMaskedValue     = "•••••"
	treeTokenExpired    = "expired"
	treeTokenValid      = "valid"
)

type treeListArguments struct {
	Depth    json.RawMessage `json:"depth"`
	ParentId string          `json:"parent_id"`
	Cursor   string          `json:"cursor"`
}

type treeItemArguments struct {
	Id string `json:"id"`
}

type treeSearchArguments struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type treeTokenArguments struct {
	OwnerId string `json:"owner_id"`
}

type treeEntry struct {
	node      dto.TreeNode
	example   *dto.TreeExample
	requestId string
	path      string
}

type treeIndex struct {
	roots   []dto.TreeNode
	entries map[string]treeEntry
	order   []string
}

type treeLine struct {
	Id       string     `json:"id"`
	Kind     string     `json:"kind"`
	Name     string     `json:"name"`
	Method   string     `json:"method,omitempty"`
	Path     string     `json:"path,omitempty"`
	Children int        `json:"children,omitempty"`
	Items    []treeLine `json:"items,omitempty"`
}

type treeRequestView struct {
	Scripts     *dto.Scripts   `json:"scripts,omitempty"`
	Auth        dto.Auth       `json:"auth"`
	Body        dto.Body       `json:"body"`
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Method      string         `json:"method"`
	Url         string         `json:"url"`
	Description string         `json:"description,omitempty"`
	Path        string         `json:"path"`
	Params      []dto.KeyValue `json:"params"`
	Headers     []dto.KeyValue `json:"headers"`
}

type treeContainerView struct {
	Id          string `json:"id"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
	Children    int    `json:"children"`
}

type treeExampleView struct {
	Id       string                  `json:"id"`
	Name     string                  `json:"name"`
	Request  dto.ExampleRequest      `json:"request"`
	Response treeExampleResponseView `json:"response"`
}

type treeExampleResponseView struct {
	Body         *string        `json:"body,omitempty"`
	StatusText   string         `json:"status_text,omitempty"`
	ContentType  string         `json:"content_type,omitempty"`
	BodyEncoding string         `json:"body_encoding,omitempty"`
	Headers      []dto.KeyValue `json:"headers,omitempty"`
	Status       int            `json:"status,omitempty"`
	Size         int            `json:"size,omitempty"`
	Binary       bool           `json:"binary,omitempty"`
}

type treeSettingsView struct {
	Scripts     *dto.Scripts       `json:"scripts,omitempty"`
	Auth        dto.Auth           `json:"auth"`
	Description string             `json:"description,omitempty"`
	Headers     []dto.KeyValue     `json:"headers"`
	Variables   []treeVariableLine `json:"variables,omitempty"`
}

type treeVariableLine struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Key         string `json:"key"`
	Initial     string `json:"initial,omitempty"`
	Current     string `json:"current,omitempty"`
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
}

type treeTokenView struct {
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Type      string     `json:"type"`
	State     string     `json:"state"`
	Scopes    []string   `json:"scopes,omitempty"`
}

func (t *tools) tree(ctx context.Context, call Call) (Result, error) {
	switch call.Action {
	case "list":
		return t.treeList(ctx, call)
	case "get":
		return t.treeGet(ctx, call)
	case "settings":
		return t.treeSettings(ctx, call)
	case "search":
		return t.treeSearch(ctx, call)
	default:
		return t.treeToken(ctx, call)
	}
}

func (t *tools) treeList(ctx context.Context, call Call) (Result, error) {
	var arguments treeListArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	depth, err := treeDepth(call.Tool, arguments.Depth)
	if err != nil {
		return Result{}, err
	}
	index, err := t.treeIndex(ctx, call.LocalId)
	if err != nil {
		return Result{}, err
	}
	nodes, examples := index.roots, []dto.TreeExample{}
	target := dto.AIActivityTarget{}
	if arguments.ParentId != "" {
		parent, found := index.entries[arguments.ParentId]
		if !found || parent.example != nil {
			return Result{}, itemNotFound(arguments.ParentId)
		}
		nodes, examples = parent.node.Children, parent.node.Examples
		target = parent.target()
	}
	page, err := Paginate(call.Tool, newTreeLines(nodes, examples, depth), arguments.Cursor)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: page, Target: target}, nil
}

func (t *tools) treeGet(ctx context.Context, call Call) (Result, error) {
	entry, err := t.treeItem(ctx, call)
	if err != nil {
		return Result{}, err
	}
	switch {
	case entry.example != nil:
		return t.treeExample(ctx, call, entry)
	case entry.isRequest():
		return t.treeRequest(ctx, call, entry)
	default:
		return t.treeContainer(ctx, call, entry)
	}
}

func (t *tools) treeRequest(ctx context.Context, call Call, entry treeEntry) (Result, error) {
	request, err := t.controllers.RequestGet.Get(ctx, dto.OpenRequest{LocalId: call.LocalId, Id: entry.node.Id})
	if err != nil {
		return Result{}, err
	}
	view := treeRequestView{
		Id:          request.Id,
		Name:        request.Name,
		Method:      request.Method,
		Url:         request.Url,
		Params:      request.Params,
		Headers:     request.Headers,
		Body:        request.Body,
		Auth:        request.Auth,
		Scripts:     request.Scripts,
		Description: request.Description,
		Path:        entry.path,
	}
	return Result{Payload: view, Target: entry.target()}, nil
}

func (t *tools) treeContainer(ctx context.Context, call Call, entry treeEntry) (Result, error) {
	settings, err := t.controllers.TreeSettings.GetSettings(ctx, dto.TreeItemReference{LocalId: call.LocalId, Id: entry.node.Id})
	if err != nil {
		return Result{}, err
	}
	view := treeContainerView{Id: entry.node.Id, Kind: entry.node.Kind, Name: entry.node.Name, Description: settings.Description, Path: entry.path, Children: entry.children()}
	return Result{Payload: view, Target: entry.target()}, nil
}

func (t *tools) treeExample(ctx context.Context, call Call, entry treeEntry) (Result, error) {
	example, err := t.controllers.ExampleManage.Get(ctx, dto.ExampleReference{LocalId: call.LocalId, RequestId: entry.requestId, ExampleId: entry.example.Id})
	if err != nil {
		return Result{}, err
	}
	view := treeExampleView{Id: example.Id, Name: example.Name, Request: example.Request, Response: newTreeExampleResponseView(example.Response)}
	return Result{Payload: view, Target: entry.target()}, nil
}

func (t *tools) treeSettings(ctx context.Context, call Call) (Result, error) {
	entry, err := t.treeItem(ctx, call)
	if err != nil {
		return Result{}, err
	}
	if !entry.isContainer() {
		return Result{}, invalidArguments(call.Tool, "settings belong to a collection or a folder")
	}
	settings, err := t.controllers.TreeSettings.GetSettings(ctx, dto.TreeItemReference{LocalId: call.LocalId, Id: entry.node.Id})
	if err != nil {
		return Result{}, err
	}
	view := treeSettingsView{Auth: settings.Auth, Headers: settings.Headers, Scripts: settings.Scripts, Description: settings.Description}
	if entry.node.Kind == itemkind.Collection.String() {
		table, err := t.controllers.VariableTable.GetTable(ctx, dto.VariableTableReference{LocalId: call.LocalId, Scope: treeCollectionScope + entry.node.Id})
		if err != nil {
			return Result{}, err
		}
		view.Variables = newTreeVariableLines(table.Rows)
	}
	return Result{Payload: view, Target: entry.target()}, nil
}

func (t *tools) treeSearch(ctx context.Context, call Call) (Result, error) {
	var arguments treeSearchArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(arguments.Query) == "" {
		return Result{}, invalidArguments(call.Tool, "query is required")
	}
	if arguments.Limit < 0 || arguments.Limit > PageSize {
		return Result{}, invalidArguments(call.Tool, "limit must be between 1 and "+strconv.Itoa(PageSize))
	}
	search, err := t.controllers.TreeSearch.Search(ctx, dto.SearchTree{LocalId: call.LocalId, Query: arguments.Query})
	if err != nil {
		return Result{}, err
	}
	index, err := t.treeIndex(ctx, call.LocalId)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: List[treeLine]{Items: index.matches(arguments.Query, search.MatchIds, treeLimit(arguments.Limit))}}, nil
}

func (t *tools) treeToken(ctx context.Context, call Call) (Result, error) {
	var arguments treeTokenArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	if arguments.OwnerId == "" {
		return Result{}, invalidArguments(call.Tool, "owner_id is required")
	}
	index, err := t.treeIndex(ctx, call.LocalId)
	if err != nil {
		return Result{}, err
	}
	owner, found := index.entries[arguments.OwnerId]
	if !found {
		return Result{}, itemNotFound(arguments.OwnerId)
	}
	token, err := t.controllers.OAuth.Get(ctx, dto.OAuthTokenReference{LocalId: call.LocalId, OwnerId: arguments.OwnerId})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newTreeTokenView(*token), Target: owner.target()}, nil
}

func (t *tools) treeItem(ctx context.Context, call Call) (treeEntry, error) {
	var arguments treeItemArguments
	if err := call.Decode(&arguments); err != nil {
		return treeEntry{}, err
	}
	if arguments.Id == "" {
		return treeEntry{}, invalidArguments(call.Tool, "id is required")
	}
	index, err := t.treeIndex(ctx, call.LocalId)
	if err != nil {
		return treeEntry{}, err
	}
	entry, found := index.entries[arguments.Id]
	if !found {
		return treeEntry{}, itemNotFound(arguments.Id)
	}
	return entry, nil
}

func (t *tools) treeIndex(ctx context.Context, localId string) (treeIndex, error) {
	tree, err := t.controllers.TreeSearch.GetTree(ctx, dto.GetTree{LocalId: localId})
	if err != nil {
		return treeIndex{}, err
	}
	index := treeIndex{roots: tree, entries: map[string]treeEntry{}}
	index.add(tree, "")
	return index, nil
}

func (i *treeIndex) add(nodes []dto.TreeNode, parentPath string) {
	for _, node := range nodes {
		path := joinTreePath(parentPath, node.Name)
		i.put(node.Id, treeEntry{node: node, path: path})
		for _, example := range node.Examples {
			i.put(example.Id, treeEntry{node: node, example: &example, requestId: node.Id, path: joinTreePath(path, example.Name)})
		}
		i.add(node.Children, path)
	}
}

func (i *treeIndex) put(id string, entry treeEntry) {
	i.entries[id] = entry
	i.order = append(i.order, id)
}

// matches keeps the tree order of the searched requests plus every collection, folder and example whose name contains the query.
func (i *treeIndex) matches(query string, ids []string, limit int) []treeLine {
	matched := make(map[string]bool, len(ids))
	for _, id := range ids {
		matched[id] = true
	}
	needle := strings.ToLower(strings.TrimSpace(query))
	lines := make([]treeLine, 0, min(len(i.order), limit))
	for _, id := range i.order {
		if len(lines) == limit {
			break
		}
		if matched[id] || strings.Contains(strings.ToLower(i.entries[id].name()), needle) {
			lines = append(lines, i.entries[id].searchLine())
		}
	}
	return lines
}

func (e treeEntry) isRequest() bool {
	return e.example == nil && e.node.Kind == itemkind.Request.String()
}

func (e treeEntry) isContainer() bool {
	return e.example == nil && (e.node.Kind == itemkind.Collection.String() || e.node.Kind == itemkind.Folder.String())
}

func (e treeEntry) name() string {
	if e.example != nil {
		return e.example.Name
	}
	return e.node.Name
}

func (e treeEntry) children() int {
	return len(e.node.Children) + len(e.node.Examples)
}

func (e treeEntry) target() dto.AIActivityTarget {
	if e.example != nil {
		return dto.AIActivityTarget{Id: e.example.Id, Name: e.example.Name, Path: e.path}
	}
	return dto.AIActivityTarget{Id: e.node.Id, Name: e.node.Name, Path: e.path}
}

func (e treeEntry) searchLine() treeLine {
	if e.example != nil {
		return treeLine{Id: e.example.Id, Kind: e.example.Kind, Name: e.example.Name, Path: e.path}
	}
	return treeLine{Id: e.node.Id, Kind: e.node.Kind, Name: e.node.Name, Method: treeMethod(e.node), Path: e.path}
}

func newTreeLines(nodes []dto.TreeNode, examples []dto.TreeExample, depth int) []treeLine {
	lines := make([]treeLine, 0, len(nodes)+len(examples))
	for _, node := range nodes {
		line := treeLine{Id: node.Id, Kind: node.Kind, Name: node.Name, Method: treeMethod(node), Children: len(node.Children) + len(node.Examples)}
		if depth == treeUnlimitedDepth || depth > 1 {
			line.Items = newTreeLines(node.Children, node.Examples, deeper(depth))
		}
		lines = append(lines, line)
	}
	for _, example := range examples {
		lines = append(lines, treeLine{Id: example.Id, Kind: example.Kind, Name: example.Name})
	}
	return lines
}

func newTreeExampleResponseView(response dto.ExampleResponse) treeExampleResponseView {
	return treeExampleResponseView{
		Body:         response.Body,
		Status:       response.Code,
		StatusText:   response.Status,
		ContentType:  response.ContentType,
		BodyEncoding: response.BodyEncoding,
		Headers:      response.Headers,
		Size:         response.SizeBytes,
		Binary:       response.IsBinary,
	}
}

func newTreeVariableLines(rows []dto.VariableRow) []treeVariableLine {
	lines := make([]treeVariableLine, 0, len(rows))
	for _, row := range rows {
		line := treeVariableLine{Key: row.Key, Description: row.Description, Initial: row.InitialValue}
		if row.HasCurrentOverride {
			line.Current = row.CurrentValue
		}
		if !row.Enabled {
			disabled := false
			line.Enabled = &disabled
		}
		if row.Secret {
			line.Secret, line.State, line.Initial, line.Current = true, row.SecretState, treeMaskedValue, ""
		}
		lines = append(lines, line)
	}
	return lines
}

func newTreeTokenView(token dto.TokenInfo) treeTokenView {
	view := treeTokenView{ExpiresAt: token.ExpiresAt, Type: token.TokenType, State: treeTokenValid, Scopes: strings.Fields(token.Scope)}
	if token.Expired {
		view.State = treeTokenExpired
	}
	return view
}

func treeDepth(tool string, raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 1, nil
	}
	var all string
	if json.Unmarshal(raw, &all) == nil && all == treeAllDepth {
		return treeUnlimitedDepth, nil
	}
	var depth int
	if err := json.Unmarshal(raw, &depth); err != nil || depth < 1 {
		return 0, invalidArguments(tool, "depth must be an integer of 1 or more, or \"all\"")
	}
	return depth, nil
}

func deeper(depth int) int {
	if depth == treeUnlimitedDepth {
		return depth
	}
	return depth - 1
}

func treeLimit(limit int) int {
	if limit == 0 {
		return treeSearchLimit
	}
	return limit
}

func treeMethod(node dto.TreeNode) string {
	if node.Method == nil {
		return ""
	}
	return *node.Method
}

func joinTreePath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + treePathSeparator + name
}
