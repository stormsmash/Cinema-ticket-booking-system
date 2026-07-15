package admin

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

type BookingFilter struct {
	Movie    string
	Status   domain.BookingStatus
	Page     int64
	PageSize int64
}

type AuditFilter struct {
	Event    domain.AuditEvent
	Page     int64
	PageSize int64
}

type BookingItem struct {
	Booking   domain.Booking
	User      domain.User
	Screening domain.Screening
}

type BookingPage struct {
	Items []BookingItem
	Total int64
}

type AuditPage struct {
	Items []domain.AuditLog
	Total int64
}

type MongoRepository struct {
	bookings   *mongo.Collection
	users      *mongo.Collection
	screenings *mongo.Collection
	auditLogs  *mongo.Collection
}

func NewMongoRepository(
	bookings *mongo.Collection,
	users *mongo.Collection,
	screenings *mongo.Collection,
	auditLogs *mongo.Collection,
) *MongoRepository {
	return &MongoRepository{
		bookings:   bookings,
		users:      users,
		screenings: screenings,
		auditLogs:  auditLogs,
	}
}

func (repository *MongoRepository) ListBookings(
	ctx context.Context,
	filter BookingFilter,
) (BookingPage, error) {
	mongoFilter := bson.D{}
	if filter.Status != "" {
		mongoFilter = append(mongoFilter, bson.E{Key: "status", Value: filter.Status})
	}
	if filter.Movie != "" {
		screeningIDs, err := repository.screeningIDsByMovie(ctx, filter.Movie)
		if err != nil {
			return BookingPage{}, err
		}
		if len(screeningIDs) == 0 {
			return BookingPage{Items: []BookingItem{}}, nil
		}
		mongoFilter = append(mongoFilter, bson.E{
			Key:   "screening_id",
			Value: bson.D{{Key: "$in", Value: screeningIDs}},
		})
	}

	total, err := repository.bookings.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return BookingPage{}, fmt.Errorf("count admin bookings: %w", err)
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip((filter.Page - 1) * filter.PageSize).
		SetLimit(filter.PageSize)
	cursor, err := repository.bookings.Find(ctx, mongoFilter, findOptions)
	if err != nil {
		return BookingPage{}, fmt.Errorf("find admin bookings: %w", err)
	}
	defer cursor.Close(ctx)

	var bookings []domain.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return BookingPage{}, fmt.Errorf("decode admin bookings: %w", err)
	}

	users, err := repository.usersByBooking(ctx, bookings)
	if err != nil {
		return BookingPage{}, err
	}
	screenings, err := repository.screeningsByBooking(ctx, bookings)
	if err != nil {
		return BookingPage{}, err
	}

	items := make([]BookingItem, 0, len(bookings))
	for _, booking := range bookings {
		items = append(items, BookingItem{
			Booking:   booking,
			User:      users[booking.UserID],
			Screening: screenings[booking.ScreeningID],
		})
	}

	return BookingPage{Items: items, Total: total}, nil
}

func (repository *MongoRepository) ListAuditLogs(
	ctx context.Context,
	filter AuditFilter,
) (AuditPage, error) {
	mongoFilter := bson.D{}
	if filter.Event != "" {
		mongoFilter = append(mongoFilter, bson.E{Key: "event", Value: filter.Event})
	}

	total, err := repository.auditLogs.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return AuditPage{}, fmt.Errorf("count audit logs: %w", err)
	}
	cursor, err := repository.auditLogs.Find(
		ctx,
		mongoFilter,
		options.Find().
			SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}).
			SetSkip((filter.Page-1)*filter.PageSize).
			SetLimit(filter.PageSize),
	)
	if err != nil {
		return AuditPage{}, fmt.Errorf("find audit logs: %w", err)
	}
	defer cursor.Close(ctx)

	var items []domain.AuditLog
	if err := cursor.All(ctx, &items); err != nil {
		return AuditPage{}, fmt.Errorf("decode audit logs: %w", err)
	}
	if items == nil {
		items = []domain.AuditLog{}
	}

	return AuditPage{Items: items, Total: total}, nil
}

func (repository *MongoRepository) screeningIDsByMovie(
	ctx context.Context,
	movie string,
) ([]bson.ObjectID, error) {
	cursor, err := repository.screenings.Find(
		ctx,
		bson.D{{Key: "movie.title", Value: movie}},
		options.Find().
			SetProjection(bson.D{{Key: "_id", Value: 1}}).
			SetCollation(&options.Collation{Locale: "en", Strength: 2}),
	)
	if err != nil {
		return nil, fmt.Errorf("find screenings by movie: %w", err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		ID bson.ObjectID `bson:"_id"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("decode screening IDs: %w", err)
	}

	ids := make([]bson.ObjectID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func (repository *MongoRepository) usersByBooking(
	ctx context.Context,
	bookings []domain.Booking,
) (map[string]domain.User, error) {
	ids := make([]bson.ObjectID, 0, len(bookings))
	seen := make(map[bson.ObjectID]struct{})
	for _, booking := range bookings {
		id, err := bson.ObjectIDFromHex(booking.UserID)
		if err != nil {
			continue
		}
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	result := make(map[string]domain.User, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	cursor, err := repository.users.Find(ctx, bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})
	if err != nil {
		return nil, fmt.Errorf("find booking users: %w", err)
	}
	defer cursor.Close(ctx)
	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("decode booking users: %w", err)
	}
	for _, user := range users {
		result[user.ID.Hex()] = user
	}
	return result, nil
}

func (repository *MongoRepository) screeningsByBooking(
	ctx context.Context,
	bookings []domain.Booking,
) (map[bson.ObjectID]domain.Screening, error) {
	ids := make([]bson.ObjectID, 0, len(bookings))
	seen := make(map[bson.ObjectID]struct{})
	for _, booking := range bookings {
		if _, exists := seen[booking.ScreeningID]; !exists {
			seen[booking.ScreeningID] = struct{}{}
			ids = append(ids, booking.ScreeningID)
		}
	}
	result := make(map[bson.ObjectID]domain.Screening, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	cursor, err := repository.screenings.Find(ctx, bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})
	if err != nil {
		return nil, fmt.Errorf("find booking screenings: %w", err)
	}
	defer cursor.Close(ctx)
	var screenings []domain.Screening
	if err := cursor.All(ctx, &screenings); err != nil {
		return nil, fmt.Errorf("decode booking screenings: %w", err)
	}
	for _, screening := range screenings {
		result[screening.ID] = screening
	}
	return result, nil
}
