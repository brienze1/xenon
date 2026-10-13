package steps

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func (t *testUtils) registerThenGitSteps(ctx *godog.ScenarioContext) {
	ctx.Then(`^the remote "([^"]*)" history should not contain "([^"]*)"$`, t.theRemoteHistoryShouldNotContain)
	ctx.Then(`^the remote "([^"]*)" branch "([^"]*)" should have (\d+) commits?$`, t.theRemoteBranchShouldHaveCommits)
	ctx.Then(`^the file "([^"]*)" on the remote "([^"]*)" should equal$`, t.theFileOnTheRemoteShouldEqual)
	ctx.Then(`^the git http server should have received (\d+) requests? with the username "([^"]*)"$`, t.theGitHttpServerShouldHaveReceivedRequestsWithTheUsername)
	ctx.Then(`^the file "([^"]*)" on the remote "([^"]*)" should (not )?exist$`, t.theFileOnTheRemoteShouldExist)
	ctx.Then(`^the last commit on the remote "([^"]*)" branch "([^"]*)" should have the (message|author) containing "([^"]*)"$`, t.theLastCommitOnTheRemoteShouldHaveTheContaining)
}

func (t *testUtils) theFileOnTheRemoteShouldExist(path, name, negation string) error {
	_, err := t.git.ReadFile(t.substitute(name), "main", t.substitute(path))
	switch {
	case negation == "" && err != nil:
		return err
	case negation != "" && err == nil:
		return fmt.Errorf("expected the remote %s not to have the file %s", name, path)
	}
	return nil
}

func (t *testUtils) theLastCommitOnTheRemoteShouldHaveTheContaining(name, branch, part, expected string) error {
	commit, err := t.git.LastCommit(t.substitute(name), t.substitute(branch))
	if err != nil {
		return err
	}
	actual := commit.Message
	if part == "author" {
		actual = commit.Author.Name + " <" + commit.Author.Email + ">"
	}
	if !strings.Contains(actual, t.substitute(expected)) {
		return fmt.Errorf("the last commit %s of %s: expected it to contain %q, got %q", part, name, t.substitute(expected), actual)
	}
	return nil
}

func (t *testUtils) theRemoteHistoryShouldNotContain(name, text string) error {
	found, err := t.git.HistoryContains(t.substitute(name), t.substitute(text))
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("the history of the remote %s contains %q", name, t.substitute(text))
	}
	return nil
}

func (t *testUtils) theRemoteBranchShouldHaveCommits(name, branch string, expected int) error {
	count, err := t.git.CommitCount(t.substitute(name), t.substitute(branch))
	if err != nil {
		return err
	}
	return expectEqual(fmt.Sprint(count), fmt.Sprint(expected), "commits on "+name+" "+branch)
}

func (t *testUtils) theFileOnTheRemoteShouldEqual(path, name string, expected *godog.DocString) error {
	content, err := t.git.ReadFile(t.substitute(name), "main", t.substitute(path))
	if err != nil {
		return err
	}
	return expectEqual(strings.TrimRight(content, "\n"), strings.TrimRight(t.substitute(expected.Content), "\n"), "remote file "+path)
}

func (t *testUtils) theGitHttpServerShouldHaveReceivedRequestsWithTheUsername(expected int, username string) error {
	count := 0
	for _, request := range t.git.HttpRequests() {
		if request.Username == t.substitute(username) {
			count++
		}
	}
	return expectEqual(fmt.Sprint(count), fmt.Sprint(expected), "git http requests with the username "+username)
}
