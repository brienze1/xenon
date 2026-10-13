package steps

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenOAuthSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the oauth provider should have received (\d+) "([^"]*)" token requests?$`, t.theOAuthProviderShouldHaveReceivedTokenRequests)
	ctx.Then(`^the (\d+) "([^"]*)" token request should have the (header|form field|query) "([^"]*)" equal to "([^"]*)"$`, t.theTokenRequestShouldHaveTheEqualTo)
	ctx.Then(`^the (\d+) "([^"]*)" token request should have the (header|form field|query) "([^"]*)" matching "([^"]*)"$`, t.theTokenRequestShouldHaveTheMatching)
	ctx.Then(`^the (\d+) "([^"]*)" token request should not have the (header|form field|query) "([^"]*)"$`, t.theTokenRequestShouldNotHaveThe)
	ctx.Then(`^the (\d+) oauth provider authorization request should have the query "([^"]*)" equal to "([^"]*)"$`, t.theAuthorizationRequestShouldHaveTheQueryEqualTo)
	ctx.Then(`^the (\d+) oauth provider authorization request should not have the query "([^"]*)"$`, t.theAuthorizationRequestShouldNotHaveTheQuery)
	ctx.Then(`^the system browser should have opened (\d+) pages?$`, t.theSystemBrowserShouldHaveOpenedPages)
	ctx.Then(`^the (\d+) page opened in the system browser should start with "([^"]*)"$`, t.thePageOpenedInTheSystemBrowserShouldStartWith)
}

func (t *testUtils) theOAuthProviderShouldHaveReceivedTokenRequests(count int, grantType string) error {
	if got := len(t.oauth.TokenRequests(grantType)); got != count {
		return fmt.Errorf("expected %d %s token requests, got %d", count, grantType, got)
	}
	return nil
}

func (t *testUtils) tokenRequestPart(index int, grantType, part, key string) (string, bool, error) {
	requests := t.oauth.TokenRequests(grantType)
	if index < 1 || index > len(requests) {
		return "", false, fmt.Errorf("there is no %s token request number %d (got %d)", grantType, index, len(requests))
	}
	request := requests[index-1]
	switch part {
	case "header":
		return request.Header.Get(key), len(request.Header.Values(key)) > 0, nil
	case "query":
		return request.Query.Get(key), request.Query.Has(key), nil
	default:
		return request.Form.Get(key), request.Form.Has(key), nil
	}
}

func (t *testUtils) theTokenRequestShouldHaveTheEqualTo(index int, grantType, part, key, expected string) error {
	value, found, err := t.tokenRequestPart(index, grantType, part, key)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("the %s token request %d has no %s %s", grantType, index, part, key)
	}
	return expectEqual(value, t.substitute(expected), part+" "+key)
}

func (t *testUtils) theTokenRequestShouldHaveTheMatching(index int, grantType, part, key, pattern string) error {
	value, found, err := t.tokenRequestPart(index, grantType, part, key)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("the %s token request %d has no %s %s", grantType, index, part, key)
	}
	return expectMatch(value, pattern, part+" "+key)
}

func (t *testUtils) theTokenRequestShouldNotHaveThe(index int, grantType, part, key string) error {
	value, found, err := t.tokenRequestPart(index, grantType, part, key)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("the %s token request %d has the %s %s = %q", grantType, index, part, key, value)
	}
	return nil
}

func (t *testUtils) authorizationRequest(index int) (url.Values, error) {
	authorizations := t.oauth.Authorizations()
	if index < 1 || index > len(authorizations) {
		return nil, fmt.Errorf("there is no authorization request number %d (got %d)", index, len(authorizations))
	}
	return authorizations[index-1], nil
}

func (t *testUtils) theAuthorizationRequestShouldHaveTheQueryEqualTo(index int, key, expected string) error {
	query, err := t.authorizationRequest(index)
	if err != nil {
		return err
	}
	if !query.Has(key) {
		return fmt.Errorf("the authorization request %d has no query %s: %v", index, key, query)
	}
	return expectEqual(query.Get(key), t.substitute(expected), "query "+key)
}

func (t *testUtils) theAuthorizationRequestShouldNotHaveTheQuery(index int, key string) error {
	query, err := t.authorizationRequest(index)
	if err != nil {
		return err
	}
	if query.Has(key) {
		return fmt.Errorf("the authorization request %d has the query %s = %q", index, key, query.Get(key))
	}
	return nil
}

func (t *testUtils) theSystemBrowserShouldHaveOpenedPages(count int) error {
	if opened := t.browser.Opened(); len(opened) != count {
		return fmt.Errorf("expected %d pages opened in the system browser, got %d: %v", count, len(opened), opened)
	}
	return nil
}

func (t *testUtils) thePageOpenedInTheSystemBrowserShouldStartWith(index int, prefix string) error {
	opened := t.browser.Opened()
	if index < 1 || index > len(opened) {
		return fmt.Errorf("there is no page number %d opened in the system browser (got %d)", index, len(opened))
	}
	if !strings.HasPrefix(opened[index-1], t.substitute(prefix)) {
		return fmt.Errorf("page %d is %q, expected it to start with %q", index, opened[index-1], t.substitute(prefix))
	}
	return nil
}
