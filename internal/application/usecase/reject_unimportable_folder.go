package usecase

import (
	"net/http"

	"github.com/brienze1/xenon/pkg/errs"
)

const (
	unimportableFolderMsg = " has no file Xenon can import; pass a Postman v3 workspace or collection folder, a Postman data dump (archive.json), a Bruno collection (bruno.json or opencollection.yml), a Xenon workspace folder, or a folder of standalone export files"
)

func RejectUnimportableFolder(name string) error {
	return errs.ClientError(name+unimportableFolderMsg, "IMP-55415", http.StatusUnsupportedMediaType)
}
