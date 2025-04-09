package service

import (
	"context"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/httperr"
	"net/http"
)

func (s *sessionService) RetrieveSession(ctx context.Context, sessionID string, sessionSecret []byte) (*model.Session, error) {
	session, err := s.sessionRepository.FindSessionByIDAndSecret(ctx, sessionID, sessionSecret)
	if err != nil {
		return nil, httperr.ErrWithStatus(http.StatusUnauthorized, "invalid session", err)
	}

	if session.IsExpired() {
		refreshedSession, err := s.RefreshSession(ctx, session)
		if err != nil {
			return nil, err
		}

		return refreshedSession, nil
	}

	return session, nil
}
