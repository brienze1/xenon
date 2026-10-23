package steps

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/cucumber/godog"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	postmanSchemaResource = "postman-collection-v2.1.0.schema.json"
	postmanSchemaUrl      = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
	zipExtension          = ".zip"
)

var (
	postmanSchemaOnce     sync.Once
	postmanSchema         *jsonschema.Schema
	postmanSchemaCompiled error
)

func (t *testUtils) registerThenExportFileSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the file "([^"]*)" should exist$`, t.theFileShouldExist)
	ctx.Then(`^the file "([^"]*)" should not exist$`, t.theFileShouldNotExist)
	ctx.Then(`^the file "([^"]*)" should not contain "([^"]*)"$`, t.theFileShouldNotContain)
	ctx.Then(`^the file "([^"]*)" should contain "([^"]*)"$`, t.theFileShouldContain)
	ctx.Then(`^the file "([^"]*)" should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theFileShouldContainTheFieldEqualTo)
	ctx.Then(`^the file "([^"]*)" should contain the field "([^"]*)" equal to$`, t.theFileShouldContainTheFieldEqualToTheText)
	ctx.Then(`^the file "([^"]*)" should contain the field "([^"]*)" array length equal to (\d+)$`, t.theFileShouldContainTheFieldArrayLengthEqualTo)
	ctx.Then(`^the file "([^"]*)" should contain the field "([^"]*)" matching "([^"]*)"$`, t.theFileShouldContainTheFieldMatching)
	ctx.Then(`^the file "([^"]*)" should contain the field "([^"]*)" with the items "([^"]*)"$`, t.theFileShouldContainTheFieldWithTheItems)
	ctx.Then(`^the file "([^"]*)" should not contain the field "([^"]*)"$`, t.theFileShouldNotContainTheField)
	ctx.Then(`^the file "([^"]*)" should equal the Postman resource file "([^"]*)" after normalisation$`, t.theFileShouldEqualThePostmanResourceFileAfterNormalisation)
	ctx.Then(`^the file "([^"]*)" should be a valid Postman v2\.1 collection$`, t.theFileShouldBeAValidPostmanCollection)
	ctx.Then(`^the zip file "([^"]*)" should contain exactly the entries "([^"]*)"$`, t.theZipFileShouldContainExactlyTheEntries)
	ctx.Step(`^the zip file "([^"]*)" is extracted to "([^"]*)"$`, t.theZipFileIsExtractedTo)
}

func (t *testUtils) tempPath(relative string) string {
	return filepath.Join(t.tempRoot, filepath.FromSlash(t.substitute(relative)))
}

func (t *testUtils) theFileShouldExist(path string) error {
	info, err := os.Stat(t.tempPath(path))
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a folder, expected a file", path)
	}
	return nil
}

func (t *testUtils) theFileShouldNotExist(path string) error {
	if _, err := os.Stat(t.tempPath(path)); !os.IsNotExist(err) {
		return fmt.Errorf("expected %s not to exist (stat error: %v)", path, err)
	}
	return nil
}

func (t *testUtils) theFileShouldNotContain(path, text string) error {
	contents, err := fileContents(t.tempPath(path))
	if err != nil {
		return err
	}
	needle := []byte(t.substitute(text))
	for name, content := range contents {
		if bytes.Contains(content, needle) {
			return fmt.Errorf("%q was found in %s%s", text, path, name)
		}
	}
	return nil
}

func (t *testUtils) theFileShouldContain(path, text string) error {
	contents, err := fileContents(t.tempPath(path))
	if err != nil {
		return err
	}
	needle := []byte(t.substitute(text))
	for _, content := range contents {
		if bytes.Contains(content, needle) {
			return nil
		}
	}
	return fmt.Errorf("%q was not found in %s", text, path)
}

func (t *testUtils) theFileShouldContainTheFieldEqualTo(path, field, expected string) error {
	value, err := yamlFileField(t.tempPath(path), field, path)
	if err != nil {
		return err
	}
	return expectEqual(stringify(value), t.substitute(expected), field)
}

func (t *testUtils) theFileShouldContainTheFieldEqualToTheText(path, field string, expected *godog.DocString) error {
	return t.theFileShouldContainTheFieldEqualTo(path, field, expected.Content)
}

func (t *testUtils) theFileShouldContainTheFieldArrayLengthEqualTo(path, field string, length int) error {
	value, err := yamlFileField(t.tempPath(path), field, path)
	if err != nil {
		return err
	}
	return expectLength(value, length, fmt.Sprintf("field %q of %s", field, path))
}

func (t *testUtils) theFileShouldContainTheFieldMatching(path, field, pattern string) error {
	value, err := yamlFileField(t.tempPath(path), field, path)
	if err != nil {
		return err
	}
	return expectMatch(stringify(value), pattern, field)
}

func (t *testUtils) theFileShouldContainTheFieldWithTheItems(path, field, items string) error {
	value, err := yamlFileField(t.tempPath(path), field, path)
	if err != nil {
		return err
	}
	return expectItems(value, t.substitute(items), field)
}

func (t *testUtils) theFileShouldNotContainTheField(path, field string) error {
	if value, err := yamlFileField(t.tempPath(path), field, path); err == nil {
		return fmt.Errorf("field %q of %s: expected no field, got %v", field, path, stringify(value))
	}
	return nil
}

func (t *testUtils) theFileShouldEqualThePostmanResourceFileAfterNormalisation(path, resource string) error {
	actual, err := jsonDocument(t.tempPath(path))
	if err != nil {
		return err
	}
	expected, err := jsonDocument(resourcePath(resource))
	if err != nil {
		return err
	}
	differences := postmanDifferences("", normalisePostman(expected), normalisePostman(actual))
	if len(differences) > 0 {
		return fmt.Errorf("%s differs from %s after normalisation:\n%s", path, resource, limitedDifferences(differences))
	}
	return nil
}

func (t *testUtils) theFileShouldBeAValidPostmanCollection(path string) error {
	schema, err := compiledPostmanSchema()
	if err != nil {
		return err
	}
	file, err := os.Open(t.tempPath(path))
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	document, err := jsonschema.UnmarshalJSON(file)
	if err != nil {
		return fmt.Errorf("%s is not JSON: %w", path, err)
	}
	if err := schema.Validate(document); err != nil {
		return fmt.Errorf("%s is not a valid Postman v2.1 collection: %#v", path, err)
	}
	return nil
}

func (t *testUtils) theZipFileShouldContainExactlyTheEntries(path, entries string) error {
	reader, err := zip.OpenReader(t.tempPath(path))
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()
	names := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		names = append(names, file.Name)
	}
	expected := strings.Split(t.substitute(entries), ", ")
	sort.Strings(names)
	sort.Strings(expected)
	return expectEqual(strings.Join(names, ", "), strings.Join(expected, ", "), path+" entries")
}

func (t *testUtils) theZipFileIsExtractedTo(path, target string) error {
	contents, err := zipContents(t.tempPath(path))
	if err != nil {
		return err
	}
	for name, content := range contents {
		destination := filepath.Join(t.tempPath(target), filepath.FromSlash(strings.TrimPrefix(name, "!")))
		if err := writeFile(destination, string(content)); err != nil {
			return err
		}
	}
	return nil
}

func compiledPostmanSchema() (*jsonschema.Schema, error) {
	postmanSchemaOnce.Do(func() {
		file, err := os.Open(resourcePath(postmanSchemaResource))
		if err != nil {
			postmanSchemaCompiled = err
			return
		}
		defer func() { _ = file.Close() }()
		document, err := jsonschema.UnmarshalJSON(file)
		if err != nil {
			postmanSchemaCompiled = err
			return
		}
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(postmanSchemaUrl, document); err != nil {
			postmanSchemaCompiled = err
			return
		}
		postmanSchema, postmanSchemaCompiled = compiler.Compile(postmanSchemaUrl)
	})
	return postmanSchema, postmanSchemaCompiled
}

func jsonDocument(path string) (any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return decodeJson(content)
}

func fileContents(path string) (map[string][]byte, error) {
	if strings.EqualFold(filepath.Ext(path), zipExtension) {
		return zipContents(path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{"": content}, nil
}

func zipContents(path string) (map[string][]byte, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()
	contents := map[string][]byte{}
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		entry, err := file.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(entry)
		_ = entry.Close()
		if err != nil {
			return nil, err
		}
		contents["!"+file.Name] = content
	}
	return contents, nil
}
