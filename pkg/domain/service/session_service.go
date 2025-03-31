package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mittwald/api-client-go/mittwaldv2"
	generatedv2 "github.com/mittwald/api-client-go/mittwaldv2/generated/clients"
	"github.com/mittwald/api-client-go/mittwaldv2/generated/clients/userclientv2"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
)

type SessionService interface {
	InitializeSessionFromRetrievalKey(ctx context.Context, atrek, userID, instanceID string) (*model.Session, error)
	RefreshSession(ctx context.Context, session *model.Session) (*model.Session, error)
}

type sessionService struct {
	client             generatedv2.Client
	sessionRepository  repository.SessionRepository
	instanceRepository repository.ExtensionInstanceRepository
}

func NewSessionService(c generatedv2.Client, sr repository.SessionRepository, ir repository.ExtensionInstanceRepository) SessionService {
	return &sessionService{
		client:             c,
		sessionRepository:  sr,
		instanceRepository: ir,
	}
}

func (s *sessionService) InitializeSessionFromRetrievalKey(ctx context.Context, atrek, userID, instanceID string) (*model.Session, error) {
	token, refresh, exp, err := s.getAPITokenFromATREK(ctx, atrek, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting access token for user %s: %w", userID, err)
	}

	instance, err := s.instanceRepository.FindExtensionInstanceByID(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("error getting instance %s: %w", instanceID, err)
	}

	authClient, err := mittwaldv2.New(ctx, mittwaldv2.WithAccessToken(token))
	if err != nil {
		return nil, fmt.Errorf("error authenticating at API: %w", err)
	}

	req := userclientv2.GetUserRequest{UserID: userID}
	resp, _, err := authClient.User().GetUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error initializing session: %w", err)
	}

	session, err := model.NewSession()
	if err != nil {
		return nil, fmt.Errorf("error initializing session: %w", err)
	}

	session.Expires = exp
	session.Email = strPtrOr(resp.Email, "")
	session.UserID = resp.UserId
	session.FirstName = resp.Person.FirstName
	session.LastName = resp.Person.LastName
	session.AccessToken = token
	session.RefreshToken = refresh
	session.Instance = instance

	if err := s.sessionRepository.CreateSessionWithUnhashedSecret(ctx, session); err != nil {
		return nil, fmt.Errorf("error creating session: %w", err)
	}

	return &session, nil
}

func (s *sessionService) RefreshSession(ctx context.Context, session *model.Session) (*model.Session, error) {
	req := userclientv2.RefreshSessionRequest{
		Body: userclientv2.RefreshSessionRequestBody{
			RefreshToken: session.RefreshToken,
		},
	}

	resp, _, err := s.client.User().RefreshSession(ctx, req)
	if err != nil {
		return nil, err
	}

	newSession := *session
	newSession.AccessToken = resp.Token
	newSession.Expires = resp.ExpiresAt
	newSession.RefreshToken = resp.RefreshToken

	if err := s.sessionRepository.RefreshSession(ctx, newSession); err != nil {
		return nil, err
	}

	return &newSession, nil
}

func (s *sessionService) getAPITokenFromATREK(ctx context.Context, atrek, userID string) (string, string, time.Time, error) {
	req := userclientv2.AuthenticateWithAccessTokenRetrievalKeyRequest{
		Body: userclientv2.AuthenticateWithAccessTokenRetrievalKeyRequestBody{
			AccessTokenRetrievalKey: atrek,
			UserId:                  userID,
		},
	}

	resp, _, err := s.client.User().AuthenticateWithAccessTokenRetrievalKey(ctx, req)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return resp.Token, resp.RefreshToken, resp.ExpiresAt, nil
}

func strPtrOr(one *string, alt string) string {
	if one != nil {
		return *one
	}
	return alt
}
