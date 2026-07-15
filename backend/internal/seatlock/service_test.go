package seatlock

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/screening"
)

type screeningFinderStub struct {
	screening domain.Screening
	err       error
}

func (stub screeningFinderStub) FindByID(
	context.Context,
	bson.ObjectID,
) (domain.Screening, error) {
	return stub.screening, stub.err
}

type storeStub struct {
	acquiredSeatID string
	acquireError   error
	releaseError   error
}

func (stub *storeStub) Acquire(
	_ context.Context,
	screeningID string,
	seatID string,
	userID string,
	ttl time.Duration,
) (Lock, error) {
	stub.acquiredSeatID = seatID
	if stub.acquireError != nil {
		return Lock{}, stub.acquireError
	}
	return Lock{
		ScreeningID: screeningID,
		SeatID:      seatID,
		UserID:      userID,
		ExpiresAt:   time.Now().Add(ttl),
	}, nil
}

func (stub *storeStub) Current(
	context.Context,
	string,
	[]string,
) (map[string]Lock, error) {
	return map[string]Lock{}, nil
}

func (stub *storeStub) Release(context.Context, string, string, string) error {
	return stub.releaseError
}

func TestAcquireValidatesAndNormalizesSeat(t *testing.T) {
	store := &storeStub{}
	service := NewService(
		screeningFinderStub{screening: domain.Screening{
			Seats: []domain.Seat{{ID: "A1", Row: "A", Number: 1}},
		}},
		store,
		10*time.Minute,
	)

	lock, err := service.Acquire(
		context.Background(),
		bson.NewObjectID().Hex(),
		" a1 ",
		"user-1",
	)
	if err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	if lock.SeatID != "A1" || store.acquiredSeatID != "A1" {
		t.Fatalf("expected normalized seat A1, got lock=%q store=%q", lock.SeatID, store.acquiredSeatID)
	}
}

func TestAcquireRejectsUnknownSeatBeforeRedis(t *testing.T) {
	store := &storeStub{}
	service := NewService(
		screeningFinderStub{screening: domain.Screening{
			Seats: []domain.Seat{{ID: "A1"}},
		}},
		store,
		time.Minute,
	)

	_, err := service.Acquire(context.Background(), bson.NewObjectID().Hex(), "B1", "user-1")
	if !errors.Is(err, ErrSeatNotFound) {
		t.Fatalf("expected ErrSeatNotFound, got %v", err)
	}
	if store.acquiredSeatID != "" {
		t.Fatal("Redis store must not be called for an unknown seat")
	}
}

func TestAcquireMapsMissingScreening(t *testing.T) {
	service := NewService(
		screeningFinderStub{err: screening.ErrNotFound},
		&storeStub{},
		time.Minute,
	)

	_, err := service.Acquire(context.Background(), bson.NewObjectID().Hex(), "A1", "user-1")
	if !errors.Is(err, ErrScreeningNotFound) {
		t.Fatalf("expected ErrScreeningNotFound, got %v", err)
	}
}

func TestAcquirePreservesContentionError(t *testing.T) {
	service := NewService(
		screeningFinderStub{screening: domain.Screening{Seats: []domain.Seat{{ID: "A1"}}}},
		&storeStub{acquireError: ErrAlreadyLocked},
		time.Minute,
	)

	_, err := service.Acquire(context.Background(), bson.NewObjectID().Hex(), "A1", "user-1")
	if !errors.Is(err, ErrAlreadyLocked) {
		t.Fatalf("expected ErrAlreadyLocked, got %v", err)
	}
}
