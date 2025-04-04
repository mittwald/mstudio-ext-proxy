package service

import (
	"context"
	"fmt"
	"github.com/mittwald/mstudio-ext-proxy/pkg/httperr"
	"net/http"

	"github.com/mittwald/api-client-go/mittwaldv2/generated/clients/userclientv2"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
)

func (s *sessionService) RefreshSession(ctx context.Context, session *model.Session) (*model.Session, error) {
	if session.RefreshToken == "" {
		return nil, httperr.ErrWithStatus(http.StatusUnauthorized, "no refresh token", fmt.Errorf("no refresh token"))
	}

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
