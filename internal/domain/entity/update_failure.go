package entity

import (
	"errors"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/updatefailure"
)

type UpdateFailure struct {
	Cause  error
	Reason enums.UpdateFailure
	Folder string
}

func NewUpdateFailure(reason enums.UpdateFailure, cause error) *UpdateFailure {
	return &UpdateFailure{Reason: reason, Cause: cause}
}

func NewUnwritableFolderFailure(folder string, cause error) *UpdateFailure {
	failure := NewUpdateFailure(updatefailure.NotWritable, cause)
	failure.Folder = folder
	return failure
}

func UpdateFailureOf(err error, fallback enums.UpdateFailure) *UpdateFailure {
	var failure *UpdateFailure
	if errors.As(err, &failure) {
		return failure
	}
	return NewUpdateFailure(fallback, err)
}

func (u *UpdateFailure) Error() string {
	if u.Cause == nil {
		return string(u.Reason)
	}
	return string(u.Reason) + ": " + u.Cause.Error()
}

func (u *UpdateFailure) Unwrap() error {
	return u.Cause
}
