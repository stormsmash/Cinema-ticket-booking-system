package booking

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

var (
	ErrBookingNotFound   = errors.New("booking not found")
	ErrSeatAlreadyBooked = errors.New("seat is already booked")
)

type MongoRepository struct {
	client     *mongo.Client
	screenings *mongo.Collection
	bookings   *mongo.Collection
	auditLogs  *mongo.Collection
}

func NewMongoRepository(
	client *mongo.Client,
	screenings *mongo.Collection,
	bookings *mongo.Collection,
	auditLogs *mongo.Collection,
) *MongoRepository {
	return &MongoRepository{
		client:     client,
		screenings: screenings,
		bookings:   bookings,
		auditLogs:  auditLogs,
	}
}

func (repository *MongoRepository) FindBooked(
	ctx context.Context,
	screeningID bson.ObjectID,
	seatID string,
) (domain.Booking, error) {
	filter := bson.D{
		{Key: "screening_id", Value: screeningID},
		{Key: "seat_id", Value: seatID},
		{Key: "status", Value: domain.BookingStatusBooked},
	}

	var item domain.Booking
	if err := repository.bookings.FindOne(ctx, filter).Decode(&item); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Booking{}, ErrBookingNotFound
		}
		return domain.Booking{}, fmt.Errorf("find booked seat: %w", err)
	}

	return item, nil
}

func (repository *MongoRepository) ListBookedByUser(
	ctx context.Context,
	userID string,
) ([]domain.Booking, error) {
	cursor, err := repository.bookings.Find(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "status", Value: domain.BookingStatusBooked},
		},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(50),
	)
	if err != nil {
		return nil, fmt.Errorf("find user bookings: %w", err)
	}
	defer cursor.Close(ctx)

	items := make([]domain.Booking, 0)
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("decode user bookings: %w", err)
	}

	return items, nil
}

func (repository *MongoRepository) CreateBooked(
	ctx context.Context,
	item domain.Booking,
) error {
	session, err := repository.client.StartSession()
	if err != nil {
		return fmt.Errorf("start booking transaction: %w", err)
	}
	defer session.EndSession(context.Background())

	auditLog := domain.AuditLog{
		ID:          bson.NewObjectID(),
		Event:       domain.AuditEventBookingSuccess,
		BookingID:   item.ID,
		UserID:      item.UserID,
		ScreeningID: item.ScreeningID,
		SeatID:      item.SeatID,
		CreatedAt:   item.CreatedAt,
	}

	_, err = session.WithTransaction(ctx, func(transactionContext context.Context) (any, error) {
		filter := bson.D{
			{Key: "_id", Value: item.ScreeningID},
			{Key: "seats", Value: bson.D{{Key: "$elemMatch", Value: bson.D{
				{Key: "id", Value: item.SeatID},
				{Key: "status", Value: domain.SeatStatusAvailable},
			}}}},
		}
		update := bson.D{{Key: "$set", Value: bson.D{
			{Key: "seats.$.status", Value: domain.SeatStatusBooked},
			{Key: "updated_at", Value: item.UpdatedAt},
		}}}

		result, err := repository.screenings.UpdateOne(transactionContext, filter, update)
		if err != nil {
			return nil, fmt.Errorf("mark seat booked: %w", err)
		}
		if result.MatchedCount == 0 {
			return nil, ErrSeatAlreadyBooked
		}

		if _, err := repository.bookings.InsertOne(transactionContext, item); err != nil {
			return nil, fmt.Errorf("insert booking: %w", err)
		}
		if _, err := repository.auditLogs.InsertOne(transactionContext, auditLog); err != nil {
			return nil, fmt.Errorf("insert booking audit log: %w", err)
		}

		return nil, nil
	})
	if mongo.IsDuplicateKeyError(err) {
		return ErrSeatAlreadyBooked
	}
	if err != nil {
		return err
	}

	return nil
}
