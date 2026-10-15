package controller

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

func mcpToolControllers() []adapter.MCPToolController {
	return []adapter.MCPToolController{
		&mcpConsole{}, &mcpCookies{}, &mcpCookiesEdit{}, &mcpDelete{}, &mcpDocs{}, &mcpDocsEdit{}, &mcpEnv{}, &mcpEnvEdit{},
		&mcpExport{}, &mcpGit{}, &mcpGitEdit{}, &mcpGitPush{}, &mcpHistory{}, &mcpImport{}, &mcpResponse{}, &mcpScripts{},
		&mcpSend{}, &mcpSettings{}, &mcpSettingsEdit{}, &mcpSnippet{}, &mcpTree{}, &mcpTreeEdit{}, &mcpWorkspace{},
		&mcpWorkspaceEdit{},
	}
}

func TestEveryMCPToolControllerIsListed(t *testing.T) {
	receivers := mcpReceivers(t)
	listed := make([]string, 0)
	for receiver, methods := range receivers {
		if methods.hasTool {
			listed = append(listed, receiver)
		}
	}
	if len(listed) != len(mcpToolControllers()) {
		t.Errorf("found %d MCP tool controllers %v, mcpToolControllers lists %d", len(listed), listed, len(mcpToolControllers()))
	}
}

func TestEveryDecodedArgumentDTOIsRegistered(t *testing.T) {
	for receiver, methods := range mcpReceivers(t) {
		for _, decoded := range methods.decoded {
			if !slices.Contains(methods.registered, decoded) {
				t.Errorf("%s decodes dto.%s but its Tool() does not register dto.NewMCPToolArguments[dto.%s]", receiver, decoded, decoded)
			}
		}
	}
}

func TestEveryMCPToolDeclaresDescribedArguments(t *testing.T) {
	for _, controller := range mcpToolControllers() {
		tool := controller.Tool()
		properties, err := tool.ArgumentProperties()
		if err != nil {
			t.Errorf("%s: %v", tool.Name, err)
			continue
		}
		for name, property := range properties {
			if len(property.Types) == 0 || !hasOwnSentence(property.Description) {
				t.Errorf("%s.%s must have a type and a description, got %+v", tool.Name, name, property)
			}
		}
		schema := tool.InputSchema()
		for _, keyword := range []string{"oneOf", "anyOf", "allOf"} {
			if schema[keyword] != nil {
				t.Errorf("%s declares %s", tool.Name, keyword)
			}
		}
	}
}

func TestHumanOnlyAIAccessIsNotDeclared(t *testing.T) {
	tool := (&mcpSettingsEdit{}).Tool()
	properties, _ := tool.ArgumentProperties()
	if _, found := properties["ai_access"]; found {
		t.Error("settings_edit must not declare ai_access")
	}
	if _, found := properties["script_timeout_ms"]; !found {
		t.Error("settings_edit must declare script_timeout_ms")
	}
}

func hasOwnSentence(description string) bool {
	description = strings.TrimSpace(description)
	return description != "" && !strings.HasPrefix(description, "Used by ") && !strings.HasPrefix(description, "Required for ")
}

type mcpReceiverMethods struct {
	hasTool    bool
	decoded    []string
	registered []string
}

func mcpReceivers(t *testing.T) map[string]*mcpReceiverMethods {
	t.Helper()
	files, err := filepath.Glob("mcp_*.go")
	if err != nil {
		t.Fatal(err)
	}
	receivers := map[string]*mcpReceiverMethods{}
	fileSet := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fileSet, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range parsed.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Recv == nil || len(function.Recv.List) == 0 {
				continue
			}
			receiver := receiverName(function.Recv.List[0].Type)
			methods := receivers[receiver]
			if methods == nil {
				methods = &mcpReceiverMethods{}
				receivers[receiver] = methods
			}
			if function.Name.Name == "Tool" {
				methods.hasTool = true
				methods.registered = append(methods.registered, registeredArguments(function)...)
			}
			methods.decoded = append(methods.decoded, decodedArguments(function)...)
		}
	}
	return receivers
}

func receiverName(expression ast.Expr) string {
	if star, isPointer := expression.(*ast.StarExpr); isPointer {
		expression = star.X
	}
	if identifier, isIdentifier := expression.(*ast.Ident); isIdentifier {
		return identifier.Name
	}
	return ""
}

func decodedArguments(function *ast.FuncDecl) []string {
	decoded := make([]string, 0)
	ast.Inspect(function, func(node ast.Node) bool {
		spec, isValue := node.(*ast.ValueSpec)
		if !isValue || len(spec.Names) == 0 || spec.Names[0].Name != "arguments" {
			return true
		}
		if name := dtoTypeName(spec.Type); name != "" {
			decoded = append(decoded, name)
		}
		return true
	})
	return decoded
}

func registeredArguments(function *ast.FuncDecl) []string {
	registered := make([]string, 0)
	ast.Inspect(function, func(node ast.Node) bool {
		index, isIndex := node.(*ast.IndexExpr)
		if !isIndex {
			return true
		}
		if selector, isSelector := index.X.(*ast.SelectorExpr); isSelector && selector.Sel.Name == "NewMCPToolArguments" {
			if name := dtoTypeName(index.Index); name != "" {
				registered = append(registered, name)
			}
		}
		return true
	})
	return registered
}

func dtoTypeName(expression ast.Expr) string {
	selector, isSelector := expression.(*ast.SelectorExpr)
	if !isSelector {
		return ""
	}
	if pkg, isIdentifier := selector.X.(*ast.Ident); !isIdentifier || pkg.Name != "dto" {
		return ""
	}
	return selector.Sel.Name
}
