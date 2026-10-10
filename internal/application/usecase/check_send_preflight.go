package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func CheckSendPreflight(request entity.Request, context entity.VariableContext, environment *entity.Environment) entity.SendPreflight {
	return entity.SendPreflight{}.WithConfirmation(environment).WithUnavailableSecretsOf(request, context)
}
