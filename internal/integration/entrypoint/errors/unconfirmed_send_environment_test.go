package errors

import (
	"errors"
	"testing"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
)

func TestUnconfirmedSendEnvironmentReadsTheRejectedEnvironment(t *testing.T) {
	rejected := usecase.RejectUnconfirmedSend(entity.SendPreflight{ConfirmEnvironment: &entity.Environment{ID: "env-1", Name: "Prod", Color: envcolor.Red}})
	environment, unconfirmed := UnconfirmedSendEnvironment(rejected)
	if !unconfirmed || environment != (entity.AIApprovalEnvironment{ID: "env-1", Name: "Prod", Color: envcolor.Red.String()}) {
		t.Fatalf("UnconfirmedSendEnvironment() = %+v, %v", environment, unconfirmed)
	}
	for _, other := range []error{nil, errors.New("raw"), RejectDeniedAISend()} {
		if _, unconfirmed := UnconfirmedSendEnvironment(other); unconfirmed {
			t.Errorf("%v must not be an unconfirmed send", other)
		}
	}
}
