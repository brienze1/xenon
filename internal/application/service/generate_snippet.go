package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type generateSnippet struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findVault             usecase.FindVault
	generateDynamicValue  usecase.GenerateDynamicValue
	signAwsRequest        usecase.SignAwsRequest
	answerDigest          usecase.AnswerDigestChallenge
	findDigestState       usecase.FindDigestState
	generateCode          usecase.GenerateSnippet
	copyToClipboard       usecase.CopyToClipboard
	variables             requestVariables
}

func NewGenerateSnippetService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	loadWorkspace usecase.LoadWorkspace,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	findVault usecase.FindVault,
	generateDynamicValue usecase.GenerateDynamicValue,
	signAwsRequest usecase.SignAwsRequest,
	answerDigest usecase.AnswerDigestChallenge,
	findDigestState usecase.FindDigestState,
	generateCode usecase.GenerateSnippet,
	copyToClipboard usecase.CopyToClipboard,
) adapter.GenerateSnippetService {
	return &generateSnippet{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findVault:             findVault,
		generateDynamicValue:  generateDynamicValue,
		signAwsRequest:        signAwsRequest,
		answerDigest:          answerDigest,
		findDigestState:       findDigestState,
		generateCode:          generateCode,
		copyToClipboard:       copyToClipboard,
		variables: requestVariables{
			findTreeItem:       findTreeItem,
			loadWorkspace:      loadWorkspace,
			findEnvironment:    findEnvironment,
			findLocalVariables: findLocalVariables,
			findVault:          findVault,
		},
	}
}

func (s *generateSnippet) Languages(ctx context.Context) []entity.SnippetLanguageSpec {
	logger.Info(ctx, "Started", nil)

	languages := s.generateCode.Languages(ctx)

	logger.Info(ctx, "Finished", map[string]any{"result": len(languages)})
	return languages
}

func (s *generateSnippet) Generate(ctx context.Context, input entity.SnippetInput) (*entity.Snippet, error) {
	logger.Info(ctx, "Started", input)

	snippet, err := s.generate(ctx, input)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"input": input, "result": snippet})
	return snippet, nil
}

func (s *generateSnippet) generate(ctx context.Context, input entity.SnippetInput) (*entity.Snippet, error) {
	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, input.LocalID)
	if err != nil {
		return nil, err
	}
	scopes, err := s.variables.load(ctx, input.LocalID, workspace.Path, input.Request.ID)
	if err != nil {
		return nil, err
	}
	resolved, resolvedAncestors := usecase.ResolveRequestVariables(input.Request, scopes.ancestors, scopes.context, s.dynamic(ctx))
	effective := usecase.ResolveEffectiveRequest(resolved, resolvedAncestors, workspace.Path).WithAuthOwner(input.TabID)
	token, err := s.token(ctx, input.LocalID, effective)
	if err != nil {
		return nil, err
	}
	secrets := append(scopes.context.SecretValues(), token.SecretsOrNone()...)
	ctx = logger.WithRedactions(ctx, secrets...)
	effective, notices, err := s.authorize(ctx, input.LocalID, usecase.MaterialiseAuth(effective), token)
	if err != nil {
		return nil, err
	}
	secrets = append(secrets, effective.Auth.DerivedSecrets(effective, secrets)...)
	ctx = logger.WithRedactions(ctx, secrets...)
	snippet, err := s.generateCode.Generate(ctx, entity.NewSnippetRequest(effective, workspace.Path, notices), input.Options)
	if err != nil {
		return nil, err
	}
	if !input.IncludeSecrets {
		masked, _ := usecase.MaskSecrets(snippet.Text, entity.SnippetSecretForms(secrets))
		*snippet = snippet.WithText(masked)
	}
	return snippet, nil
}

func (s *generateSnippet) Copy(ctx context.Context, input entity.SnippetInput) error {
	logger.Info(ctx, "Started", input)

	snippet, err := s.generate(ctx, input)
	if err != nil {
		return err
	}
	if err := s.copyToClipboard.Copy(ctx, snippet.Text); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", input)
	return nil
}

func (s *generateSnippet) dynamic(ctx context.Context) entity.DynamicValue {
	return func(name string) (string, bool) {
		return s.generateDynamicValue.Generate(ctx, name)
	}
}

func (s *generateSnippet) token(ctx context.Context, localId string, request entity.EffectiveRequest) (*entity.OAuth2Token, error) {
	if !request.Auth.IsOAuth2() {
		return nil, nil
	}
	vault, err := s.findVault.Find(ctx, localId)
	if err != nil {
		return nil, err
	}
	token, found := vault.TokenOf(request.AuthOwner)
	if !found {
		return nil, nil
	}
	return &token, nil
}

func (s *generateSnippet) authorize(ctx context.Context, localId string, request entity.EffectiveRequest, token *entity.OAuth2Token) (entity.EffectiveRequest, []string, error) {
	auth := request.Auth
	switch {
	case auth.IsApiKey():
		return usecase.ApplyApiKey(request, auth.ApiKey), nil, nil
	case auth.IsOAuth2():
		return usecase.ApplyOAuthToken(request, auth.OAuth2, token), entity.OAuthSnippetNotices(token), nil
	case auth.IsAwsV4():
		signed, err := s.signAwsRequest.Sign(ctx, request, auth.AwsV4)
		if err != nil {
			return request, nil, err
		}
		return *signed, nil, nil
	case auth.IsDigest():
		return s.digest(ctx, localId, request)
	case auth.IsUnsupported():
		return request, []string{entity.UnsupportedAuthNotice(auth)}, nil
	default:
		return request, nil, nil
	}
}

func (s *generateSnippet) digest(ctx context.Context, localId string, request entity.EffectiveRequest) (entity.EffectiveRequest, []string, error) {
	stored, err := s.findDigestState.FindDigestState(ctx, localId, request.AuthOwner)
	if err != nil {
		return request, nil, err
	}
	state := request.Auth.Digest.KnownState(stored)
	if state == nil {
		return request, []string{entity.DigestChallengeNotice}, nil
	}
	header, err := s.answerDigest.Answer(ctx, request, request.Auth.Digest, state.Next())
	if err != nil {
		return request, nil, err
	}
	return request.WithHeader(entity.HeaderAuthorization, header), nil, nil
}
