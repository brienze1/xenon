package usecase

import (
	"net/http"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	gitNextStepAttribute    = "next_step"
	gitFingerprintAttribute = "fingerprint"
)

func RejectMissingGitAccount(id string) error {
	return errs.NotFoundError("no Git account with the id "+id, "ACC-04404")
}

func RejectUnverifiableToken() error {
	return errs.BadRequestError("Xenon cannot verify a token for this host on its own; enter a repository the token can read", "ACC-03400")
}

func RejectForeignRepository(repository, host string) error {
	return errs.BadRequestError("the repository "+repository+" is not on the account's host "+host, "ACC-18400")
}

func RejectUnknownHostKey(key entity.HostKey) error {
	err := errs.BadRequestError("the SSH host "+key.Host+" is not known yet; confirm its key fingerprint", "ACC-05409",
		errs.ErrorDetails{Attribute: gitFingerprintAttribute, Messages: []string{key.Fingerprint}},
		errs.ErrorDetails{Attribute: gitNextStepAttribute, Messages: []string{"Compare the fingerprint with the one your Git host publishes, then confirm it"}})
	err.SetStatusCode(http.StatusConflict, err.Description())
	return err
}

func RejectChangedHostKey(key entity.HostKey) error {
	err := errs.BadRequestError("the SSH host key of "+key.Host+" changed; the connection was blocked", "ACC-06409",
		errs.ErrorDetails{Attribute: gitFingerprintAttribute, Messages: []string{key.Fingerprint}},
		errs.ErrorDetails{Attribute: gitNextStepAttribute, Messages: []string{"Ask the host's administrator whether the key was rotated before trusting it"}})
	err.SetStatusCode(http.StatusConflict, err.Description())
	return err
}

func RejectHostKeyMismatch(key entity.HostKey) error {
	return errs.BadRequestError("the host "+key.Host+" now presents the fingerprint "+key.Fingerprint+"; nothing was remembered", "ACC-15400")
}

func RejectNotSshHost(repository string) error {
	return errs.BadRequestError("the repository "+repository+" is not reached over SSH", "ACC-19400")
}

func RejectPublicKeyUpload(reason string) error {
	return errs.BadRequestError("the public key cannot be uploaded: "+reason, "ACC-12400")
}
