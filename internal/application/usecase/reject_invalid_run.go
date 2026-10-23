package usecase

import "github.com/brienze1/xenon/pkg/errs"

const (
	runIdAttribute  = "run_id"
	reasonAttribute = "reason"
)

func RejectUnknownRun(runId string) error {
	return errs.NotFoundError("no active or stored run with the id "+runId, "RUN-01404")
}

func RejectActiveRun(runId string) error {
	return errs.ConflictError("a run is already active on this collection or folder", "RUN-02409",
		errs.ErrorDetails{Attribute: runIdAttribute, Messages: []string{runId}})
}

func RejectInvalidRunSetting(field, reason string) error {
	return errs.BadRequestError("the run setting "+field+" "+reason, "RUN-04400",
		errs.ErrorDetails{Attribute: field, Messages: []string{reason}})
}

func RejectInvalidIterationData(reason string) error {
	return errs.BadRequestError("the iteration data file cannot be used: "+reason, "RUN-05400",
		errs.ErrorDetails{Attribute: reasonAttribute, Messages: []string{reason}})
}

func NewMissingRunRequest(requestId string) error {
	return errs.NotFoundError("request not found: "+requestId, "RUN-06404")
}

func RejectUnknownRunTarget(itemId string) error {
	return errs.NotFoundError("no collection or folder with the id "+itemId, "RUN-07404")
}
