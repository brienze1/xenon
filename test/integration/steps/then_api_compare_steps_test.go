package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/cucumber/godog"
)

type receivedForm struct {
	values map[string][]string
	files  map[string]string
}

func (t *testUtils) registerThenApiCompareSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the (\d+) and (\d+) "([^"]*)" requests to "([^"]*)" should have the same query, body and headers "([^"]*)"$`, t.theRequestsShouldHaveTheSameQueryBodyAndHeaders)
}

func (t *testUtils) theRequestsShouldHaveTheSameQueryBodyAndHeaders(first, second int, method, path, headers string) error {
	firstHeaders, firstQuery, firstBody, found := t.api.GetRequest(first, method, t.substitute(path))
	if !found {
		return fmt.Errorf("request %d %s %s was not received", first, method, path)
	}
	secondHeaders, secondQuery, secondBody, found := t.api.GetRequest(second, method, t.substitute(path))
	if !found {
		return fmt.Errorf("request %d %s %s was not received", second, method, path)
	}
	if !reflect.DeepEqual(firstQuery, secondQuery) {
		return fmt.Errorf("%s %s: query %v differs from %v", method, path, firstQuery, secondQuery)
	}
	for _, name := range strings.Split(headers, ", ") {
		if name == "" {
			continue
		}
		if expected, actual := strings.Join(firstHeaders.Values(name), ","), strings.Join(secondHeaders.Values(name), ","); expected != actual || expected == "" {
			return fmt.Errorf("%s %s: header %s %q differs from %q", method, path, name, expected, actual)
		}
	}
	return sameBody(firstHeaders, firstBody, secondHeaders, secondBody)
}

func sameBody(firstHeaders http.Header, first []byte, secondHeaders http.Header, second []byte) error {
	mediaType, _, _ := mime.ParseMediaType(firstHeaders.Get("Content-Type"))
	switch mediaType {
	case "multipart/form-data":
		expected, err := multipartForm(firstHeaders, first)
		if err != nil {
			return err
		}
		actual, err := multipartForm(secondHeaders, second)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("form %+v differs from %+v", expected, actual)
		}
	case "application/x-www-form-urlencoded":
		expected, _ := url.ParseQuery(string(first))
		actual, _ := url.ParseQuery(string(second))
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("form %v differs from %v", expected, actual)
		}
	case "application/json":
		var expected, actual any
		if json.Unmarshal(first, &expected) != nil || json.Unmarshal(second, &actual) != nil || !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("JSON body %s differs from %s", first, second)
		}
	default:
		if !bytes.Equal(first, second) {
			return fmt.Errorf("body %q differs from %q", first, second)
		}
	}
	return nil
}

func multipartForm(headers http.Header, body []byte) (receivedForm, error) {
	_, params, err := mime.ParseMediaType(headers.Get("Content-Type"))
	if err != nil {
		return receivedForm{}, err
	}
	reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	form := receivedForm{values: map[string][]string{}, files: map[string]string{}}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			return form, nil
		}
		if err != nil {
			return receivedForm{}, err
		}
		content, err := io.ReadAll(part)
		if err != nil {
			return receivedForm{}, err
		}
		if part.FileName() != "" {
			form.files[part.FormName()] = part.FileName() + ":" + string(content)
			continue
		}
		form.values[part.FormName()] = append(form.values[part.FormName()], string(content))
	}
}
