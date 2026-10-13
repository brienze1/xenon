package usecase

import (
	"net/http"

	"github.com/brienze1/xenon/pkg/errs"
)

func RejectNotLinked(localId string) error {
	err := errs.BadRequestError("the workspace "+localId+" is not linked to a Git repository", "GIT-09409",
		errs.ErrorDetails{Attribute: gitNextStepAttribute, Messages: []string{"Link the workspace to a repository in its Sync settings"}})
	err.SetStatusCode(http.StatusConflict, err.Description())
	return err
}

func RejectMissingSyncAccount() error {
	err := errs.BadRequestError("No account - choose one", "GIT-07428",
		errs.ErrorDetails{Attribute: gitNextStepAttribute, Messages: []string{"Choose a Git account in the workspace's Sync settings"}})
	err.SetStatusCode(http.StatusPreconditionRequired, err.Description())
	return err
}

func RejectExistingWorkspace(subfolder string) error {
	return errs.ConflictError("the subfolder \""+subfolder+"\" already holds a Xenon workspace", "GIT-01409",
		errs.ErrorDetails{Attribute: gitNextStepAttribute, Messages: []string{"Choose \"Use the repo's workspace\" or \"Replace it with mine\""}})
}

func RejectLargeFileStorage(subfolder string) error {
	err := errs.BadRequestError("the repository tracks files of \""+subfolder+"\" with Git Large File Storage, which Xenon does not support", "GIT-02415",
		errs.ErrorDetails{Attribute: gitNextStepAttribute, Messages: []string{"Use a subfolder without Large File Storage, or another repository"}})
	err.SetStatusCode(http.StatusUnsupportedMediaType, err.Description())
	return err
}

func RejectClonePathInUse(clonePath string) error {
	return errs.ConflictError("the folder "+clonePath+" is not empty; pick an empty or new folder for the clone", "GIT-65409")
}

func RejectClonePathOverlap(clonePath string) error {
	return errs.ConflictError("the folder "+clonePath+" overlaps a workspace folder; pick a folder outside every workspace", "GIT-66409")
}

func RejectUnlinkTargetInUse(target string) error {
	return errs.ConflictError("the folder "+target+" is not empty; pick an empty or new folder for the workspace files", "GIT-43409")
}

func RejectUnlinkTargetOverlap(target string) error {
	return errs.ConflictError("the folder "+target+" overlaps a workspace folder or Xenon's clone; pick a folder outside both", "GIT-44409")
}

func RejectEmptyCommit() error {
	return errs.BadRequestError("none of the chosen items has changes to commit", "GIT-08400")
}

func RejectUnresolvedConflicts() error {
	return errs.ConflictError("resolve every conflict before finishing the pull", "GIT-04409",
		errs.ErrorDetails{Attribute: gitNextStepAttribute, Messages: []string{"Choose Keep mine, Keep theirs or Keep both for each remaining item"}})
}

func RejectConflictResolution(itemName, resolution, fieldPath string) error {
	return errs.BadRequestError("\""+resolution+"\" does not resolve the conflict of "+itemName+describeField(fieldPath), "GIT-19400")
}

func RejectMissingConflict(itemId string) error {
	return errs.NotFoundError("the pull has no conflict for the item "+itemId, "GIT-35404")
}

func RejectStaleSyncSession() error {
	return errs.ConflictError("the branch changed since this pull; the conflicts no longer apply", "GIT-28409",
		errs.ErrorDetails{Attribute: gitNextStepAttribute, Messages: []string{"Pull again to see the current conflicts"}})
}

func describeField(fieldPath string) string {
	if fieldPath == "" {
		return ""
	}
	return " (field \"" + fieldPath + "\")"
}
