package persistence

import (
	"context"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/model"
	"github.com/mittwald/mstudio-ext-proxy/pkg/domain/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var _ repository.ExtensionInstanceRepository = &mongoExtensionInstanceRepository{}

type mongoExtensionInstanceRepository struct {
	collection *mongo.Collection
}

func NewMongoExtensionInstanceRepository(collection *mongo.Collection) repository.ExtensionInstanceRepository {
	return &mongoExtensionInstanceRepository{
		collection: collection,
	}
}

func (m *mongoExtensionInstanceRepository) FindExtensionInstanceByID(ctx context.Context, instanceID string) (model.ExtensionInstance, error) {
	out := model.ExtensionInstance{}
	err := m.collection.FindOne(ctx, bson.M{"_id": instanceID}).Decode(&out)

	return out, err
}

func (m *mongoExtensionInstanceRepository) AddExtensionInstance(ctx context.Context, instance model.ExtensionInstance) error {
	_, err := m.collection.InsertOne(ctx, instance)
	return err
}

func (m *mongoExtensionInstanceRepository) UpdateExtensionInstance(ctx context.Context, instance model.ExtensionInstance) error {
	_, err := m.collection.ReplaceOne(ctx, bson.M{"_id": instance.ID}, instance)
	return err
}

func (m *mongoExtensionInstanceRepository) RemoveExtensionInstance(ctx context.Context, instance model.ExtensionInstance) error {
	return m.RemoveExtensionInstanceByID(ctx, instance.ID)
}

func (m *mongoExtensionInstanceRepository) RemoveExtensionInstanceByID(ctx context.Context, instanceID string) error {
	_, err := m.collection.DeleteOne(ctx, bson.M{"_id": instanceID})
	return err
}
