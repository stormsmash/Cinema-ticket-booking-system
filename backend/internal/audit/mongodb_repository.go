package audit

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(collection *mongo.Collection) *MongoRepository {
	return &MongoRepository{collection: collection}
}

func (repository *MongoRepository) Create(ctx context.Context, item domain.AuditLog) error {
	if _, err := repository.collection.InsertOne(ctx, item); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}
