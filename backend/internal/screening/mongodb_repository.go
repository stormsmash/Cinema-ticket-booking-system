package screening

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(collection *mongo.Collection) *MongoRepository {
	return &MongoRepository{collection: collection}
}

func (repository *MongoRepository) List(ctx context.Context) ([]domain.Screening, error) {
	cursor, err := repository.collection.Find(
		ctx,
		bson.D{},
		options.Find().SetSort(bson.D{{Key: "starts_at", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}
	defer cursor.Close(ctx)

	var screenings []domain.Screening
	if err := cursor.All(ctx, &screenings); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	return screenings, nil
}

func (repository *MongoRepository) FindByID(
	ctx context.Context,
	id bson.ObjectID,
) (domain.Screening, error) {
	var screening domain.Screening
	err := repository.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&screening)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Screening{}, ErrNotFound
	}
	if err != nil {
		return domain.Screening{}, fmt.Errorf("find one: %w", err)
	}

	return screening, nil
}
