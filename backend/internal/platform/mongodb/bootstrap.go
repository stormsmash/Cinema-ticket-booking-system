package mongodb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

const (
	CollectionScreenings = "screenings"
	CollectionBookings   = "bookings"
	CollectionUsers      = "users"
	CollectionAuditLogs  = "audit_logs"
)

func Bootstrap(ctx context.Context, database *mongo.Database, adminEmails []string) error {
	if err := createIndexes(ctx, database); err != nil {
		return err
	}
	if err := configureUserRoles(ctx, database.Collection(CollectionUsers), adminEmails); err != nil {
		return err
	}

	if err := seedScreenings(ctx, database.Collection(CollectionScreenings), time.Now().UTC()); err != nil {
		return err
	}
	if err := backfillSeatStatuses(ctx, database.Collection(CollectionScreenings)); err != nil {
		return err
	}

	return nil
}

func createIndexes(ctx context.Context, database *mongo.Database) error {
	_, err := database.Collection(CollectionUsers).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "google_subject", Value: 1}},
			Options: options.Index().
				SetName("unique_google_subject").
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: options.Index().
				SetName("user_email"),
		},
	})
	if err != nil {
		return fmt.Errorf("create user indexes: %w", err)
	}

	_, err = database.Collection(CollectionScreenings).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "starts_at", Value: 1}},
			Options: options.Index().
				SetName("screening_starts_at"),
		},
		{
			Keys: bson.D{{Key: "movie.title", Value: 1}},
			Options: options.Index().
				SetName("screening_movie_title"),
		},
	})
	if err != nil {
		return fmt.Errorf("create screening indexes: %w", err)
	}

	_, err = database.Collection(CollectionBookings).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "created_at", Value: -1},
				{Key: "_id", Value: -1},
			},
			Options: options.Index().SetName("booking_created_at"),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
				{Key: "_id", Value: -1},
			},
			Options: options.Index().SetName("booking_status_created_at"),
		},
		{
			Keys: bson.D{
				{Key: "screening_id", Value: 1},
				{Key: "seat_id", Value: 1},
			},
			Options: options.Index().
				SetName("unique_booked_seat").
				SetUnique(true).
				SetPartialFilterExpression(bson.D{
					{Key: "status", Value: domain.BookingStatusBooked},
				}),
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().
				SetName("booking_status_expiry"),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().
				SetName("booking_user_created_at"),
		},
	})
	if err != nil {
		return fmt.Errorf("create booking indexes: %w", err)
	}

	_, err = database.Collection(CollectionAuditLogs).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "booking_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("audit_booking_created_at"),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().
				SetName("audit_created_at"),
		},
		{
			Keys: bson.D{
				{Key: "event", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().
				SetName("audit_event_created_at"),
		},
	})
	if err != nil {
		return fmt.Errorf("create audit log indexes: %w", err)
	}

	return nil
}

func configureUserRoles(
	ctx context.Context,
	collection *mongo.Collection,
	adminEmails []string,
) error {
	missingRole := bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "role", Value: bson.D{{Key: "$exists", Value: false}}}},
		bson.D{{Key: "role", Value: nil}},
		bson.D{{Key: "role", Value: ""}},
	}}}
	if _, err := collection.UpdateMany(
		ctx,
		missingRole,
		bson.D{{Key: "$set", Value: bson.D{{Key: "role", Value: domain.UserRoleUser}}}},
	); err != nil {
		return fmt.Errorf("backfill user roles: %w", err)
	}

	if len(adminEmails) == 0 {
		return nil
	}
	if _, err := collection.UpdateMany(
		ctx,
		bson.D{{Key: "email", Value: bson.D{{Key: "$in", Value: adminEmails}}}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "role", Value: domain.UserRoleAdmin}}}},
		options.UpdateMany().SetCollation(&options.Collation{Locale: "en", Strength: 2}),
	); err != nil {
		return fmt.Errorf("promote configured admins: %w", err)
	}

	return nil
}

func backfillSeatStatuses(ctx context.Context, collection *mongo.Collection) error {
	filter := bson.D{{Key: "seats", Value: bson.D{{Key: "$elemMatch", Value: bson.D{
		{Key: "status", Value: bson.D{{Key: "$exists", Value: false}}},
	}}}}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "seats.$[seat].status", Value: domain.SeatStatusAvailable},
	}}}
	updateOptions := options.UpdateMany().SetArrayFilters([]any{
		bson.D{{Key: "seat.status", Value: bson.D{{Key: "$exists", Value: false}}}},
	})

	if _, err := collection.UpdateMany(ctx, filter, update, updateOptions); err != nil {
		return fmt.Errorf("backfill screening seat statuses: %w", err)
	}

	return nil
}

func seedScreenings(ctx context.Context, collection *mongo.Collection, now time.Time) error {
	for _, screening := range screeningSeeds(now) {
		filter := bson.D{{Key: "_id", Value: screening.ID}}
		update := bson.D{{
			Key: "$setOnInsert",
			Value: bson.D{
				{Key: "movie", Value: screening.Movie},
				{Key: "auditorium", Value: screening.Auditorium},
				{Key: "starts_at", Value: screening.StartsAt},
				{Key: "seats", Value: screening.Seats},
				{Key: "created_at", Value: screening.CreatedAt},
				{Key: "updated_at", Value: screening.UpdatedAt},
			},
		}}

		if _, err := collection.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
			return fmt.Errorf("seed screening %s: %w", screening.ID.Hex(), err)
		}
	}

	return nil
}

func screeningSeeds(now time.Time) []domain.Screening {
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	createdAt := now.Truncate(time.Second)

	return []domain.Screening{
		{
			ID: mustObjectID("66a000000000000000000001"),
			Movie: domain.Movie{
				Title:           "Midnight Signal",
				DurationMinutes: 118,
			},
			Auditorium: domain.Auditorium{
				Name:        "Hall 1",
				Rows:        5,
				SeatsPerRow: 10,
			},
			StartsAt:  startOfDay.Add(24*time.Hour + 19*time.Hour),
			Seats:     buildSeats([]string{"A", "B", "C", "D", "E"}, 10),
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
		{
			ID: mustObjectID("66a000000000000000000002"),
			Movie: domain.Movie{
				Title:           "The Last Orbit",
				DurationMinutes: 132,
			},
			Auditorium: domain.Auditorium{
				Name:        "Hall 2",
				Rows:        4,
				SeatsPerRow: 8,
			},
			StartsAt:  startOfDay.Add(24*time.Hour + 21*time.Hour),
			Seats:     buildSeats([]string{"A", "B", "C", "D"}, 8),
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
	}
}

func buildSeats(rows []string, seatsPerRow int) []domain.Seat {
	seats := make([]domain.Seat, 0, len(rows)*seatsPerRow)
	for _, row := range rows {
		for number := 1; number <= seatsPerRow; number++ {
			seats = append(seats, domain.Seat{
				ID:     row + strconv.Itoa(number),
				Row:    row,
				Number: number,
				Status: domain.SeatStatusAvailable,
			})
		}
	}

	return seats
}

func mustObjectID(value string) bson.ObjectID {
	id, err := bson.ObjectIDFromHex(value)
	if err != nil {
		panic(fmt.Sprintf("invalid seed ObjectID %q: %v", value, err))
	}

	return id
}
