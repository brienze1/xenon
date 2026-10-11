package steps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/brienze1/xenon/test/integration/mock"
)

func (t *testUtils) registerGivenGitSteps(ctx *godog.ScenarioContext) {
	ctx.Given(`^the remote "([^"]*)" exists$`, t.theRemoteExists)
	ctx.Given(`^the remote "([^"]*)" has the file "([^"]*)" with$`, t.theRemoteHasTheFileWith)
	ctx.Step(`^a teammate changes "([^"]*)" on the remote "([^"]*)" to$`, t.aTeammateChangesOnTheRemoteTo)
	ctx.Given(`^the "([^"]*)" variable is set to the (http|ssh) url of the remote "([^"]*)"$`, t.theVariableIsSetToTheUrlOfTheRemote)
	ctx.Given(`^the "([^"]*)" variable is set to the git (http|ssh) server host$`, t.theVariableIsSetToTheGitServerHost)
	ctx.Given(`^the "([^"]*)" variable is set to the git ssh server (known hosts line|host key fingerprint)$`, t.theVariableIsSetToTheGitSshServerHostKey)
	ctx.Given(`^the git http server requires the username "([^"]*)" and the token "([^"]*)"$`, t.theGitHttpServerRequiresTheUsernameAndTheToken)
	ctx.Given(`^the git http server requires any username and the token "([^"]*)"$`, t.theGitHttpServerRequiresAnyUsernameAndTheToken)
	ctx.Given(`^the git http server answers the remote "([^"]*)" with status (\d+) and the body "([^"]*)"$`, t.theGitHttpServerAnswersTheRemoteWithStatusAndTheBody)
	ctx.Given(`^an ssh key "([^"]*)" is saved at the temp path "([^"]*)"(?: with the passphrase "([^"]*)")?$`, t.anSshKeyIsSavedAtTheTempPath)
	ctx.Given(`^the git ssh server authorizes the ssh key "([^"]*)"$`, t.theGitSshServerAuthorizesTheSshKey)
	ctx.Step(`^the git ssh server authorizes the public key "([^"]*)"$`, t.theGitSshServerAuthorizesThePublicKey)
	ctx.Given(`^the ssh agent holds the ssh key "([^"]*)"$`, t.theSshAgentHoldsTheSshKey)
	ctx.Given(`^the ssh agent is running without keys$`, t.theSshAgentIsRunningWithoutKeys)
}

func (t *testUtils) theRemoteExists(name string) error {
	return t.git.CreateRemote(t.substitute(name))
}

func (t *testUtils) theRemoteHasTheFileWith(name, path string, content *godog.DocString) error {
	return t.git.CommitFile(t.substitute(name), t.substitute(path), t.substitute(content.Content)+"\n", "Add "+path)
}

func (t *testUtils) aTeammateChangesOnTheRemoteTo(path, name string, content *godog.DocString) error {
	return t.git.CommitFile(t.substitute(name), t.substitute(path), t.substitute(content.Content)+"\n", "Change "+path)
}

func (t *testUtils) theVariableIsSetToTheUrlOfTheRemote(variable, protocol, name string) error {
	if protocol == "ssh" {
		t.vars[variable] = t.git.SshURL(t.substitute(name))
		return nil
	}
	t.vars[variable] = t.git.HttpURL(t.substitute(name))
	return nil
}

func (t *testUtils) theVariableIsSetToTheGitServerHost(variable, protocol string) error {
	if protocol == "ssh" {
		t.vars[variable] = knownhosts.Normalize(t.git.SshHost())
		return nil
	}
	t.vars[variable] = t.git.HttpHost()
	return nil
}

func (t *testUtils) theVariableIsSetToTheGitSshServerHostKey(variable, what string) error {
	key := t.git.SshHostKey()
	if what == "host key fingerprint" {
		t.vars[variable] = gossh.FingerprintSHA256(key)
		return nil
	}
	t.vars[variable] = knownhosts.Line([]string{t.git.SshHost()}, key)
	return nil
}

func (t *testUtils) theGitHttpServerRequiresTheUsernameAndTheToken(username, token string) error {
	t.git.RequireHttpCredentials(t.substitute(username), t.substitute(token), false)
	return nil
}

func (t *testUtils) theGitHttpServerRequiresAnyUsernameAndTheToken(token string) error {
	t.git.RequireHttpCredentials("", t.substitute(token), true)
	return nil
}

func (t *testUtils) theGitHttpServerAnswersTheRemoteWithStatusAndTheBody(name string, status int, body string) error {
	t.git.FailHttp(t.substitute(name), status, t.substitute(body))
	return nil
}

func (t *testUtils) anSshKeyIsSavedAtTheTempPath(name, path, passphrase string) error {
	key, err := mock.GenerateSshKey(t.substitute(passphrase))
	if err != nil {
		return err
	}
	location := filepath.Join(t.tempRoot, filepath.FromSlash(t.substitute(path)))
	if err := os.MkdirAll(filepath.Dir(location), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(location, key.PrivatePem, 0o600); err != nil {
		return err
	}
	t.sshKeys[name] = key
	t.vars[name+"_public_key"] = strings.TrimSpace(key.AuthorizedKey())
	t.vars[name+"_fingerprint"] = gossh.FingerprintSHA256(key.Signer.PublicKey())
	return nil
}

func (t *testUtils) theGitSshServerAuthorizesTheSshKey(name string) error {
	key, found := t.sshKeys[name]
	if !found {
		return fmt.Errorf("no ssh key named %s", name)
	}
	t.git.AuthorizeSshKey(key.Signer.PublicKey())
	return nil
}

func (t *testUtils) theGitSshServerAuthorizesThePublicKey(authorized string) error {
	key, _, _, _, err := gossh.ParseAuthorizedKey([]byte(t.substitute(authorized)))
	if err != nil {
		return fmt.Errorf("not an authorized key %q: %w", t.substitute(authorized), err)
	}
	t.git.AuthorizeSshKey(key)
	return nil
}

func (t *testUtils) theSshAgentHoldsTheSshKey(name string) error {
	key, found := t.sshKeys[name]
	if !found {
		return fmt.Errorf("no ssh key named %s", name)
	}
	return t.sshAgent.Hold(key.Private)
}

func (t *testUtils) theSshAgentIsRunningWithoutKeys() error {
	t.sshAgent.Start()
	return nil
}
