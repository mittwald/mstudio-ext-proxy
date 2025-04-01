package service

import (
	"context"
	generatedv2 "github.com/mittwald/api-client-go/mittwaldv2/generated/clients"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
)

type SessionService interface {
	InitializeSessionFromRetrievalKey(ctx context.Context, atrek, userID, instanceID string) (*model.Session, error)
	RetrieveSession(ctx context.Context, sessionID string, sessionSecret []byte) (*model.Session, error)
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
