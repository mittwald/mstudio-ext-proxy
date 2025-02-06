package repository

import (
	"context"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
)

type SessionRepository interface {
	FindSessionByIDAndSecret(ctx context.Context, id string, secret []byte) (*model.Session, error)
	CreateSession(ctx context.Context, session model.Session) error
	CreateSessionWithUnhashedSecret(ctx context.Context, session model.Session) error
}
