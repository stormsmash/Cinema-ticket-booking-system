package booking

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/realtime"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/seatlock"
)

type repositoryStub struct {
	existing    domain.Booking
	findError   error
	createError error
	created     domain.Booking
}

func (stub *repositoryStub) FindBooked(
	context.Context,
	bson.ObjectID,
	string,
) (domain.Booking, error) {
	return stub.existing, stub.findError
}

func (stub *repositoryStub) CreateBooked(_ context.Context, item domain.Booking) error {
	stub.created = item
	return stub.createError
}

type lockClaimerStub struct {
	claimError error
	claim      seatlock.Claim
	claimed    bool
	restored   bool
	committed  bool
}

func (stub *lockClaimerStub) Claim(
	context.Context,
	string,
	string,
	string,
	time.Duration,
) (seatlock.Claim, error) {
	stub.claimed = true
	return stub.claim, stub.claimError
}

func (stub *lockClaimerStub) RestoreClaim(context.Context, seatlock.Claim) error {
	stub.restored = true
	return nil
}

func (stub *lockClaimerStub) CommitClaim(context.Context, seatlock.Claim) error {
	stub.committed = true
	return nil
}

type eventPublisherStub struct {
	events []realtime.SeatEvent
	err    error
}

func (stub *eventPublisherStub) Publish(_ context.Context, event realtime.SeatEvent) error {
	stub.events = append(stub.events, event)
	return stub.err
}

func bookingTestScreening() domain.Screening {
	return domain.Screening{
		ID:       bson.NewObjectID(),
		StartsAt: time.Now().UTC().Add(time.Hour),
		Seats: []domain.Seat{{
			ID:     "A1",
			Status: domain.SeatStatusAvailable,
		}},
	}
}

func TestConfirmCreatesBookingThenCommitsClaim(t *testing.T) {
	item := bookingTestScreening()
	repository := &repositoryStub{findError: ErrBookingNotFound}
	locks := &lockClaimerStub{claim: seatlock.Claim{SeatID: "A1"}}
	events := &eventPublisherStub{}
	service := NewService(repository, screeningFinderStub{screening: item}, locks, events)

	confirmation, err := service.Confirm(context.Background(), item.ID.Hex(), " a1 ", "user-1")
	if err != nil {
		t.Fatalf("confirm booking: %v", err)
	}
	if !confirmation.Created || repository.created.Status != domain.BookingStatusBooked {
		t.Fatalf("expected a new booked record, got %#v", confirmation)
	}
	if repository.created.SeatID != "A1" || repository.created.UserID != "user-1" {
		t.Fatalf("unexpected booking: %#v", repository.created)
	}
	if !locks.claimed || !locks.committed || locks.restored {
		t.Fatalf("unexpected claim lifecycle: %#v", locks)
	}
	if len(events.events) != 1 || events.events[0].Type != realtime.SeatBooked ||
		events.events[0].BookingID != confirmation.Booking.ID.Hex() {
		t.Fatalf("expected one booked event, got %#v", events.events)
	}
}

func TestConfirmKeepsBookingWhenRealtimePublishFails(t *testing.T) {
	item := bookingTestScreening()
	repository := &repositoryStub{findError: ErrBookingNotFound}
	locks := &lockClaimerStub{}
	events := &eventPublisherStub{err: errors.New("Redis unavailable")}
	service := NewService(repository, screeningFinderStub{screening: item}, locks, events)

	confirmation, err := service.Confirm(context.Background(), item.ID.Hex(), "A1", "user-1")
	if err != nil {
		t.Fatalf("confirm booking: %v", err)
	}
	if !confirmation.Created || len(events.events) != 1 {
		t.Fatalf("booking must stay successful after publish failure: %#v", confirmation)
	}
	if !locks.committed || locks.restored {
		t.Fatalf("committed booking claim must not be restored: %#v", locks)
	}
}

func TestConfirmRejectsExpiredLockBeforeMongoWrite(t *testing.T) {
	item := bookingTestScreening()
	repository := &repositoryStub{findError: ErrBookingNotFound}
	locks := &lockClaimerStub{claimError: seatlock.ErrLockNotFound}
	service := NewService(repository, screeningFinderStub{screening: item}, locks, nil)

	_, err := service.Confirm(context.Background(), item.ID.Hex(), "A1", "user-1")
	if !errors.Is(err, ErrSeatLockExpired) {
		t.Fatalf("expected ErrSeatLockExpired, got %v", err)
	}
	if !repository.created.ID.IsZero() {
		t.Fatal("MongoDB must not be called without an owned lock")
	}
}

func TestConfirmRestoresClaimWhenMongoFails(t *testing.T) {
	item := bookingTestScreening()
	repository := &repositoryStub{
		findError:   ErrBookingNotFound,
		createError: errors.New("MongoDB unavailable"),
	}
	locks := &lockClaimerStub{}
	events := &eventPublisherStub{}
	service := NewService(repository, screeningFinderStub{screening: item}, locks, events)

	if _, err := service.Confirm(context.Background(), item.ID.Hex(), "A1", "user-1"); err == nil {
		t.Fatal("expected MongoDB failure")
	}
	if !locks.restored || locks.committed {
		t.Fatalf("expected claim restoration, got %#v", locks)
	}
	if len(events.events) != 0 {
		t.Fatalf("failed booking must not publish an event: %#v", events.events)
	}
}

func TestConfirmReturnsExistingBookingForSameUser(t *testing.T) {
	item := bookingTestScreening()
	existing := domain.Booking{
		ID:          bson.NewObjectID(),
		UserID:      "user-1",
		ScreeningID: item.ID,
		SeatID:      "A1",
		Status:      domain.BookingStatusBooked,
	}
	repository := &repositoryStub{existing: existing}
	locks := &lockClaimerStub{}
	events := &eventPublisherStub{}
	service := NewService(repository, screeningFinderStub{screening: item}, locks, events)

	confirmation, err := service.Confirm(context.Background(), item.ID.Hex(), "A1", "user-1")
	if err != nil {
		t.Fatalf("repeat confirmation: %v", err)
	}
	if confirmation.Created || confirmation.Booking.ID != existing.ID {
		t.Fatalf("expected existing booking, got %#v", confirmation)
	}
	if locks.claimed {
		t.Fatal("idempotent retry must not require the deleted Redis lock")
	}
	if len(events.events) != 0 {
		t.Fatal("idempotent retry must not publish a duplicate event")
	}
}

func TestConfirmRejectsStartedScreening(t *testing.T) {
	item := bookingTestScreening()
	item.StartsAt = time.Now().UTC().Add(-time.Minute)
	locks := &lockClaimerStub{}
	service := NewService(
		&repositoryStub{findError: ErrBookingNotFound},
		screeningFinderStub{screening: item},
		locks,
		nil,
	)

	_, err := service.Confirm(context.Background(), item.ID.Hex(), "A1", "user-1")
	if !errors.Is(err, ErrScreeningStarted) {
		t.Fatalf("expected ErrScreeningStarted, got %v", err)
	}
	if locks.claimed {
		t.Fatal("started screening must not claim a Redis lock")
	}
}

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
