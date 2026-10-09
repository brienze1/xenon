package steps

import (
	"time"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerGivenApiSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the (\d+) "([^"]*)" request to "([^"]*)" returns status (\d+) with the following response$`, t.theRequestReturnsStatusWithTheFollowingResponse)
	ctx.Given(`^the (\d+) "([^"]*)" request to "([^"]*)" returns status (\d+) with content type "([^"]*)" and body$`, t.theRequestReturnsStatusWithContentTypeAndBody)
	ctx.Given(`^the "([^"]*)" requests to "([^"]*)" respond after (\d+)ms$`, t.theRequestsRespondAfter)
	ctx.Given(`^the requests to "([^"]*)" redirect to "([^"]*)" with status (\d+)$`, t.theRequestsRedirectTo)
}

func (t *testUtils) theRequestReturnsStatusWithTheFollowingResponse(index int, method, path string, status int, body *godog.DocString) error {
	t.api.SetResponse(index, method, path, status, []byte(t.substitute(body.Content)), "application/json")
	return nil
}

func (t *testUtils) theRequestReturnsStatusWithContentTypeAndBody(index int, method, path string, status int, contentType string, body *godog.DocString) error {
	t.api.SetResponse(index, method, path, status, []byte(t.substitute(body.Content)), contentType)
	return nil
}

func (t *testUtils) theRequestsRespondAfter(method, path string, milliseconds int) error {
	t.api.SetDelay(method, path, time.Duration(milliseconds)*time.Millisecond)
	return nil
}

func (t *testUtils) theRequestsRedirectTo(path, target string, status int) error {
	t.api.SetRedirect(path, t.substitute(target), status)
	return nil
}
