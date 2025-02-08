package persistence

import (
	"context"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var _ repository.SessionRepository = &mongoSessionRepository{}

type mongoSessionRepository struct {
	collection *mongo.Collection
}

func MustNewMongoSessionRepository(collection *mongo.Collection) repository.SessionRepository {
	repo, err := NewMongoSessionRepository(collection)
	if err != nil {
		panic(err)
	}

	return repo
}

func NewMongoSessionRepository(collection *mongo.Collection) (repository.SessionRepository, error) {
	repo := &mongoSessionRepository{
		collection: collection,
	}

	if err := repo.Setup(context.Background()); err != nil {
		return nil, err
	}

	return repo, nil
}

func (m *mongoSessionRepository) Setup(ctx context.Context) error {
	_, err := m.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	return err
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
