package service

import (
	"context"
	mittwaldv2 "github.com/mittwald/api-client-go/mittwaldv2/generated/clients"
	"github.com/mittwald/api-client-go/mittwaldv2/generated/clients/userclientv2"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
)

type SessionService interface {
	RefreshSession(ctx context.Context, session *model.Session) (*model.Session, error)
}

type sessionService struct {
	client mittwaldv2.Client
	repo   repository.SessionRepository
}

func NewSessionService(c mittwaldv2.Client, r repository.SessionRepository) SessionService {
	return &sessionService{
		client: c,
		repo:   r,
	}
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

	if err := s.repo.RefreshSession(ctx, newSession); err != nil {
		return nil, err
	}

	return &newSession, nil
}
