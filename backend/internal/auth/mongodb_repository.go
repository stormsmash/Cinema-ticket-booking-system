package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(collection *mongo.Collection) *MongoUserRepository {
	return &MongoUserRepository{collection: collection}
}

func (repository *MongoUserRepository) UpsertGoogleUser(
	ctx context.Context,
	profile GoogleProfile,
) (domain.User, error) {
	now := time.Now().UTC()
	filter := bson.D{{Key: "google_subject", Value: profile.Subject}}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "email", Value: profile.Email},
				{Key: "name", Value: profile.Name},
				{Key: "avatar_url", Value: profile.AvatarURL},
				{Key: "updated_at", Value: now},
			},
		},
		{
			Key: "$setOnInsert",
			Value: bson.D{
				{Key: "google_subject", Value: profile.Subject},
				{Key: "created_at", Value: now},
			},
		},
	}

	var user domain.User
	err := repository.collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&user)
	if err != nil {
		return domain.User{}, fmt.Errorf("upsert: %w", err)
	}

	return user, nil
}

func (repository *MongoUserRepository) FindByID(
	ctx context.Context,
	userID string,
) (domain.User, error) {
	id, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return domain.User{}, ErrUserNotFound
	}

	var user domain.User
	err = repository.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.User{}, ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find one: %w", err)
	}

	return user, nil
}
