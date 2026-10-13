package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenDigestSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the digest server should have received (\d+) requests? and accepted (\d+)$`, t.theDigestServerShouldHaveReceivedRequestsAndAccepted)
	ctx.Then(`^the (\d+) request to the digest server should have the header "Authorization" matching "([^"]*)"$`, t.theRequestToTheDigestServerShouldHaveTheAuthorizationMatching)
}

func (t *testUtils) theDigestServerShouldHaveReceivedRequestsAndAccepted(received, accepted int) error {
	gotReceived, gotAccepted := t.digest.Counts()
	if gotReceived != received || gotAccepted != accepted {
		return fmt.Errorf("the digest server received %d and accepted %d requests, expected %d and %d", gotReceived, gotAccepted, received, accepted)
	}
	return nil
}

func (t *testUtils) theRequestToTheDigestServerShouldHaveTheAuthorizationMatching(index int, pattern string) error {
	authorization, found := t.digest.Authorization(index)
	if !found {
		return fmt.Errorf("the digest server has no request number %d", index)
	}
	return expectMatch(authorization, pattern, "header Authorization")
}
