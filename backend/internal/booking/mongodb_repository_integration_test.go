package booking

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	mongostore "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/platform/mongodb"
)

func TestMongoRepositoryPreventsConcurrentDoubleBooking(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("set MONGO_TEST_URI to run the MongoDB concurrency test")
	}

	setupContext, cancelSetup := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelSetup()
	client, err := mongostore.Connect(setupContext, uri)
	if err != nil {
		t.Fatalf("connect to test MongoDB: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	database := client.Database(fmt.Sprintf("cinema_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		cleanupContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = database.Drop(cleanupContext)
	})
	if err := mongostore.Bootstrap(setupContext, database, nil); err != nil {
		t.Fatalf("bootstrap test database: %v", err)
	}

	screeningID := bson.NewObjectID()
	now := time.Now().UTC()
	screening := domain.Screening{
		ID:         screeningID,
		Movie:      domain.Movie{Title: "Concurrency Test", DurationMinutes: 90},
		Auditorium: domain.Auditorium{Name: "Test Screen", Rows: 1, SeatsPerRow: 1},
		StartsAt:   now.Add(time.Hour),
		Seats: []domain.Seat{{
			ID:     "A1",
			Row:    "A",
			Number: 1,
			Status: domain.SeatStatusAvailable,
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := database.Collection(mongostore.CollectionScreenings).InsertOne(setupContext, screening); err != nil {
		t.Fatalf("insert test screening: %v", err)
	}

	repository := NewMongoRepository(
		client,
		database.Collection(mongostore.CollectionScreenings),
		database.Collection(mongostore.CollectionBookings),
		database.Collection(mongostore.CollectionAuditLogs),
	)
	requests := []domain.Booking{
		newConcurrentTestBooking(screeningID, "user-1", now),
		newConcurrentTestBooking(screeningID, "user-2", now),
	}
	results := make(chan error, len(requests))
	start := make(chan struct{})
	var waitGroup sync.WaitGroup
	for _, item := range requests {
		item := item
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			operationContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			results <- repository.CreateBooked(operationContext, item)
		}()
	}

	close(start)
	waitGroup.Wait()
	close(results)

	successCount := 0
	conflictCount := 0
	for result := range results {
		switch {
		case result == nil:
			successCount++
		case errors.Is(result, ErrSeatAlreadyBooked):
			conflictCount++
		default:
			t.Fatalf("unexpected booking result: %v", result)
		}
	}
	if successCount != 1 || conflictCount != 1 {
		t.Fatalf("expected one booking and one conflict, got success=%d conflict=%d", successCount, conflictCount)
	}

	bookingCount, err := database.Collection(mongostore.CollectionBookings).CountDocuments(
		setupContext,
		bson.D{{Key: "screening_id", Value: screeningID}, {Key: "seat_id", Value: "A1"}},
	)
	if err != nil {
		t.Fatalf("count persisted bookings: %v", err)
	}
	if bookingCount != 1 {
		t.Fatalf("expected one persisted booking, got %d", bookingCount)
	}

	auditCount, err := database.Collection(mongostore.CollectionAuditLogs).CountDocuments(
		setupContext,
		bson.D{{Key: "screening_id", Value: screeningID}, {Key: "event", Value: domain.AuditEventBookingSuccess}},
	)
	if err != nil {
		t.Fatalf("count persisted audit logs: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one booking audit log, got %d", auditCount)
	}

	var stored domain.Screening
	if err := database.Collection(mongostore.CollectionScreenings).
		FindOne(setupContext, bson.D{{Key: "_id", Value: screeningID}}).
		Decode(&stored); err != nil {
		t.Fatalf("load stored screening: %v", err)
	}
	if len(stored.Seats) != 1 || stored.Seats[0].Status != domain.SeatStatusBooked {
		t.Fatalf("expected durable BOOKED seat, got %#v", stored.Seats)
	}
}

func newConcurrentTestBooking(screeningID bson.ObjectID, userID string, now time.Time) domain.Booking {
	return domain.Booking{
		ID:          bson.NewObjectID(),
		UserID:      userID,
		ScreeningID: screeningID,
		SeatID:      "A1",
		Status:      domain.BookingStatusBooked,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
