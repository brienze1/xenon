package steps

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"
	"github.com/cucumber/godog"
	"go.yaml.in/yaml/v3"
)

var ulidPattern = regexp.MustCompile(`[0-9A-HJKMNP-TV-Z]{26}`)

func (t *testUtils) registerThenFileCompareSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the (?:file|folder) "([^"]*)" should equal the (?:file|folder) "([^"]*)"$`, t.theFileShouldEqualTheFile)
	ctx.Then(`^the (?:file|folder) "([^"]*)" should equal the (?:file|folder) "([^"]*)" apart from ids$`, t.theFileShouldEqualTheFileApartFromIds)
	ctx.Then(`^the (?:field "([^"]*)" of the )?file "([^"]*)" decrypted with the passphrase "([^"]*)" should contain the field "([^"]*)" equal to "([^"]*)"$`, t.theDecryptedFileShouldContainTheFieldEqualTo)
	ctx.Then(`^the (?:field "([^"]*)" of the )?file "([^"]*)" decrypted with the passphrase "([^"]*)" should not contain the field "([^"]*)"$`, t.theDecryptedFileShouldNotContainTheField)
}

func (t *testUtils) theFileShouldEqualTheFile(actual, expected string) error {
	return t.compareTrees(actual, expected, func(content []byte) []byte { return content })
}

func (t *testUtils) theFileShouldEqualTheFileApartFromIds(actual, expected string) error {
	return t.compareTrees(actual, expected, func(content []byte) []byte {
		return ulidPattern.ReplaceAll(content, []byte("<id>"))
	})
}

func (t *testUtils) compareTrees(actual, expected string, normalise func([]byte) []byte) error {
	actualFiles, err := treeFiles(t.tempPath(actual))
	if err != nil {
		return err
	}
	expectedFiles, err := treeFiles(t.tempPath(expected))
	if err != nil {
		return err
	}
	if err := expectEqual(strings.Join(sortedKeys(actualFiles), ", "), strings.Join(sortedKeys(expectedFiles), ", "), actual+" files"); err != nil {
		return err
	}
	for name, content := range actualFiles {
		if !bytes.Equal(normalise(content), normalise(expectedFiles[name])) {
			return fmt.Errorf("%s%s differs from %s%s:\n--- got\n%s\n--- expected\n%s", actual, name, expected, name, normalise(content), normalise(expectedFiles[name]))
		}
	}
	return nil
}

func (t *testUtils) theDecryptedFileShouldContainTheFieldEqualTo(field, path, passphrase, key, expected string) error {
	document, err := t.decryptedDocument(field, path, passphrase)
	if err != nil {
		return err
	}
	value, found := lookup(document, t.substitute(key))
	if !found {
		return fmt.Errorf("field %q not found in the decrypted secrets of %s", key, path)
	}
	return expectEqual(stringify(value), t.substitute(expected), key)
}

func (t *testUtils) theDecryptedFileShouldNotContainTheField(field, path, passphrase, key string) error {
	document, err := t.decryptedDocument(field, path, passphrase)
	if err != nil {
		return err
	}
	if value, found := lookup(document, t.substitute(key)); found {
		return fmt.Errorf("field %q of the decrypted secrets of %s: expected no field, got %v", key, path, stringify(value))
	}
	return nil
}

func (t *testUtils) decryptedDocument(field, path, passphrase string) (any, error) {
	armored, err := os.ReadFile(t.tempPath(path))
	if err != nil {
		return nil, err
	}
	if field != "" {
		value, err := yamlFileField(t.tempPath(path), field, path)
		if err != nil {
			return nil, err
		}
		armored = []byte(stringify(value))
	}
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, err
	}
	reader, err := age.Decrypt(armor.NewReader(bytes.NewReader(armored)), identity)
	if err != nil {
		return nil, fmt.Errorf("the secrets of %s cannot be decrypted: %w", path, err)
	}
	plaintext, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var document any
	if err := yaml.Unmarshal(plaintext, &document); err != nil {
		return nil, fmt.Errorf("the decrypted secrets of %s are not JSON: %w", path, err)
	}
	return document, nil
}

func treeFiles(root string) (map[string][]byte, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		content, err := os.ReadFile(root)
		return map[string][]byte{"": content}, err
	}
	files := map[string][]byte{}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		files["/"+filepath.ToSlash(relative)] = content
		return nil
	})
	return files, err
}

func sortedKeys(files map[string][]byte) []string {
	keys := make([]string, 0, len(files))
	for key := range files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
