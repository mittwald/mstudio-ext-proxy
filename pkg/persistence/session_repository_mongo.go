package persistence

import (
	"context"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var _ repository.SessionRepository = &mongoSessionRepository{}

type mongoSessionRepository struct {
	collection *mongo.Collection
}

func NewMongoSessionRepository(collection *mongo.Collection) repository.SessionRepository {
	return &mongoSessionRepository{
		collection: collection,
	}
}

func (m *mongoSessionRepository) FindSessionByIDAndSecret(ctx context.Context, id string, secret []byte) (*model.Session, error) {
	session := model.Session{}

	if err := m.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&session); err != nil {
		_, _ = bcrypt.GenerateFromPassword([]byte("dummy password"), bcrypt.DefaultCost)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(session.SessionSecret, secret); err != nil {
		return nil, err
	}

	return &session, nil
}

func (m *mongoSessionRepository) CreateSessionWithUnhashedSecret(ctx context.Context, session model.Session) error {
	enc, err := bcrypt.GenerateFromPassword(session.SessionSecret, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	session.SessionSecret = enc
	return m.CreateSession(ctx, session)
}

func (m *mongoSessionRepository) CreateSession(ctx context.Context, session model.Session) error {
	_, err := m.collection.InsertOne(ctx, session)
	return err
}
