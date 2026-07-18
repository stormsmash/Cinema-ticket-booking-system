package booking

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/realtime"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/screening"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/seatlock"
)

const (
	bookingOperationTimeout = 10 * time.Second
	bookingClaimTTL         = 15 * time.Second
	bookingCleanupTimeout   = 2 * time.Second
)

var (
	ErrInvalidScreeningID = errors.New("invalid screening ID")
	ErrScreeningNotFound  = errors.New("screening not found")
	ErrScreeningStarted   = errors.New("screening has already started")
	ErrSeatNotFound       = errors.New("seat not found")
	ErrSeatLockExpired    = errors.New("seat lock expired")
	ErrSeatLockNotOwned   = errors.New("seat lock belongs to another user")
)

type Repository interface {
	FindBooked(context.Context, bson.ObjectID, string) (domain.Booking, error)
	CreateBooked(context.Context, domain.Booking) error
	ListBookedByUser(context.Context, string) ([]domain.Booking, error)
}

type ScreeningFinder interface {
	FindByID(context.Context, bson.ObjectID) (domain.Screening, error)
}

type LockClaimer interface {
	Claim(context.Context, string, string, string, time.Duration) (seatlock.Claim, error)
	RestoreClaim(context.Context, seatlock.Claim) error
	CommitClaim(context.Context, seatlock.Claim) error
}

type EventPublisher interface {
	Publish(context.Context, realtime.SeatEvent) error
}

type Confirmation struct {
	Booking domain.Booking
	Created bool
}

type UserTicket struct {
	Booking   domain.Booking
	Screening domain.Screening
}

type Service struct {
	repository Repository
	screenings ScreeningFinder
	locks      LockClaimer
	events     EventPublisher
	now        func() time.Time
}

func NewService(
	repository Repository,
	screenings ScreeningFinder,
	locks LockClaimer,
	events EventPublisher,
) *Service {
	return &Service{
		repository: repository,
		screenings: screenings,
		locks:      locks,
		events:     events,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func (service *Service) Confirm(
	ctx context.Context,
	screeningID string,
	seatID string,
	userID string,
) (Confirmation, error) {
	operationContext, cancel := context.WithTimeout(ctx, bookingOperationTimeout)
	defer cancel()

	item, seat, err := service.validateSeat(operationContext, screeningID, seatID)
	if err != nil {
		return Confirmation{}, err
	}

	existing, err := service.repository.FindBooked(operationContext, item.ID, seat.ID)
	if err == nil {
		if existing.UserID == userID {
			return Confirmation{Booking: existing, Created: false}, nil
		}
		return Confirmation{}, ErrSeatAlreadyBooked
	}
	if !errors.Is(err, ErrBookingNotFound) {
		return Confirmation{}, fmt.Errorf("check existing booking: %w", err)
	}
	if seat.Status == domain.SeatStatusBooked {
		return Confirmation{}, ErrSeatAlreadyBooked
	}
	if !item.StartsAt.After(service.now()) {
		return Confirmation{}, ErrScreeningStarted
	}

	claim, err := service.locks.Claim(
		operationContext,
		item.ID.Hex(),
		seat.ID,
		userID,
		bookingClaimTTL,
	)
	if errors.Is(err, seatlock.ErrLockNotFound) {
		return Confirmation{}, ErrSeatLockExpired
	}
	if errors.Is(err, seatlock.ErrLockNotOwned) {
		return Confirmation{}, ErrSeatLockNotOwned
	}
	if err != nil {
		return Confirmation{}, fmt.Errorf("claim seat lock: %w", err)
	}

	now := service.now()
	created := domain.Booking{
		ID:          bson.NewObjectID(),
		UserID:      userID,
		ScreeningID: item.ID,
		SeatID:      seat.ID,
		PriceBaht:   item.TicketPriceBaht,
		Status:      domain.BookingStatusBooked,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := service.repository.CreateBooked(operationContext, created); err != nil {
		existing, found := service.findBookedAfterFailure(item.ID, seat.ID)
		if found {
			service.finishClaim(claim)
			if existing.UserID == userID {
				return Confirmation{Booking: existing, Created: false}, nil
			}
			return Confirmation{}, ErrSeatAlreadyBooked
		}
		if errors.Is(err, ErrSeatAlreadyBooked) {
			service.finishClaim(claim)
			return Confirmation{}, ErrSeatAlreadyBooked
		}

		service.restoreClaim(claim)
		return Confirmation{}, fmt.Errorf("create booking: %w", err)
	}

	service.finishClaim(claim)
	service.publishBooked(created)

	return Confirmation{Booking: created, Created: true}, nil
}

func (service *Service) ListForUser(ctx context.Context, userID string) ([]UserTicket, error) {
	operationContext, cancel := context.WithTimeout(ctx, bookingOperationTimeout)
	defer cancel()

	bookings, err := service.repository.ListBookedByUser(operationContext, userID)
	if err != nil {
		return nil, fmt.Errorf("list user bookings: %w", err)
	}

	tickets := make([]UserTicket, 0, len(bookings))
	for _, item := range bookings {
		screeningItem, err := service.screenings.FindByID(operationContext, item.ScreeningID)
		if err != nil {
			return nil, fmt.Errorf("find ticket screening: %w", err)
		}
		tickets = append(tickets, UserTicket{Booking: item, Screening: screeningItem})
	}

	return tickets, nil
}

func (service *Service) findBookedAfterFailure(
	screeningID bson.ObjectID,
	seatID string,
) (domain.Booking, bool) {
	lookupContext, cancel := context.WithTimeout(context.Background(), bookingCleanupTimeout)
	defer cancel()

	item, err := service.repository.FindBooked(lookupContext, screeningID, seatID)
	return item, err == nil
}

func (service *Service) validateSeat(
	ctx context.Context,
	screeningID string,
	seatID string,
) (domain.Screening, domain.Seat, error) {
	id, err := bson.ObjectIDFromHex(screeningID)
	if err != nil {
		return domain.Screening{}, domain.Seat{}, ErrInvalidScreeningID
	}

	item, err := service.screenings.FindByID(ctx, id)
	if errors.Is(err, screening.ErrNotFound) {
		return domain.Screening{}, domain.Seat{}, ErrScreeningNotFound
	}
	if err != nil {
		return domain.Screening{}, domain.Seat{}, fmt.Errorf("find screening: %w", err)
	}

	seatID = strings.ToUpper(strings.TrimSpace(seatID))
	for _, seat := range item.Seats {
		if seat.ID == seatID {
			if seat.Status == "" {
				seat.Status = domain.SeatStatusAvailable
			}
			return item, seat, nil
		}
	}

	return domain.Screening{}, domain.Seat{}, ErrSeatNotFound
}

func (service *Service) restoreClaim(claim seatlock.Claim) {
	cleanupContext, cancel := context.WithTimeout(context.Background(), bookingCleanupTimeout)
	defer cancel()

	if err := service.locks.RestoreClaim(cleanupContext, claim); err != nil {
		log.Printf("restore booking seat claim: %v", err)
	}
}

func (service *Service) finishClaim(claim seatlock.Claim) {
	cleanupContext, cancel := context.WithTimeout(context.Background(), bookingCleanupTimeout)
	defer cancel()

	if err := service.locks.CommitClaim(cleanupContext, claim); err != nil {
		log.Printf("finish booking seat claim: %v", err)
	}
}

func (service *Service) publishBooked(item domain.Booking) {
	if service.events == nil {
		return
	}

	publishContext, cancel := context.WithTimeout(context.Background(), bookingCleanupTimeout)
	defer cancel()
	event := realtime.SeatEvent{
		Version:     realtime.EventVersion,
		Type:        realtime.SeatBooked,
		BookingID:   item.ID.Hex(),
		ScreeningID: item.ScreeningID.Hex(),
		SeatID:      item.SeatID,
		Status:      string(domain.SeatStatusBooked),
		OccurredAt:  item.CreatedAt,
	}
	if err := service.events.Publish(publishContext, event); err != nil {
		log.Printf("publish booked seat event: %v", err)
	}
}
