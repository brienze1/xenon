package steps

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenFileSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the file "([^"]*)" should equal the resource file "([^"]*)"$`, t.theFileShouldEqualTheResourceFile)
	ctx.Then(`^the file "([^"]*)" should be the generated body of (\d+) bytes$`, t.theFileShouldBeTheGeneratedBody)
}

func (t *testUtils) theFileShouldEqualTheResourceFile(path, name string) error {
	expected, err := os.ReadFile(resourcePath(name))
	if err != nil {
		return err
	}
	return t.fileShouldHaveBytes(path, expected)
}

func (t *testUtils) theFileShouldBeTheGeneratedBody(path string, size int) error {
	return t.fileShouldHaveBytes(path, generatedBody(size))
}

func (t *testUtils) fileShouldHaveBytes(path string, expected []byte) error {
	content, err := os.ReadFile(filepath.Join(t.tempRoot, filepath.FromSlash(t.substitute(path))))
	if err != nil {
		return err
	}
	return expectSameBytes(content, expected, path)
}

func filesShouldNotContain(root, text string, skip func(path string) bool) error {
	if text == "" {
		return fmt.Errorf("the searched text must not be empty")
	}
	matches := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || skip(path) {
			return err
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if bytes.Contains(content, []byte(text)) {
			relative, _ := filepath.Rel(root, path)
			matches = append(matches, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(matches) > 0 {
		return fmt.Errorf("%q was found in %s", text, strings.Join(matches, ", "))
	}
	return nil
}
