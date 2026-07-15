package seatlock

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/screening"
)

var (
	ErrInvalidScreeningID = errors.New("invalid screening ID")
	ErrScreeningNotFound  = errors.New("screening not found")
	ErrSeatNotFound       = errors.New("seat not found")
	ErrAlreadyLocked      = errors.New("seat is already locked")
	ErrAlreadyBooked      = errors.New("seat is already booked")
	ErrLockNotFound       = errors.New("seat lock was not found")
	ErrLockNotOwned       = errors.New("seat lock belongs to another user")
)

type Lock struct {
	ScreeningID string
	SeatID      string
	UserID      string
	ExpiresAt   time.Time
}

type ScreeningFinder interface {
	FindByID(context.Context, bson.ObjectID) (domain.Screening, error)
}

type Store interface {
	Acquire(context.Context, string, string, string, time.Duration) (Lock, error)
	Current(context.Context, string, []string) (map[string]Lock, error)
	Release(context.Context, string, string, string) error
}

type AuditRecorder interface {
	Create(context.Context, domain.AuditLog) error
}

type Service struct {
	screenings ScreeningFinder
	store      Store
	audits     AuditRecorder
	ttl        time.Duration
}

func NewService(
	screenings ScreeningFinder,
	store Store,
	audits AuditRecorder,
	ttl time.Duration,
) *Service {
	return &Service{screenings: screenings, store: store, audits: audits, ttl: ttl}
}

func (service *Service) Acquire(
	ctx context.Context,
	screeningID string,
	seatID string,
	userID string,
) (Lock, error) {
	seatID, err := service.validateSeat(ctx, screeningID, seatID)
	if err != nil {
		return Lock{}, err
	}

	lock, err := service.store.Acquire(ctx, screeningID, seatID, userID, service.ttl)
	if err != nil {
		if !errors.Is(err, ErrAlreadyLocked) {
			service.recordAudit(ctx, domain.AuditLog{
				Event:       domain.AuditEventSystemError,
				UserID:      userID,
				ScreeningID: mustObjectID(screeningID),
				SeatID:      seatID,
				Message:     "acquire seat lock",
			})
		}
		return Lock{}, fmt.Errorf("acquire seat lock: %w", err)
	}

	return lock, nil
}

func (service *Service) CurrentLocks(
	ctx context.Context,
	screeningID string,
	seats []domain.Seat,
) (map[string]Lock, error) {
	seatIDs := make([]string, 0, len(seats))
	for _, seat := range seats {
		seatIDs = append(seatIDs, seat.ID)
	}

	locks, err := service.store.Current(ctx, screeningID, seatIDs)
	if err != nil {
		return nil, fmt.Errorf("load current seat locks: %w", err)
	}

	return locks, nil
}

func (service *Service) Release(
	ctx context.Context,
	screeningID string,
	seatID string,
	userID string,
) error {
	seatID, err := service.validateSeat(ctx, screeningID, seatID)
	if err != nil {
		return err
	}

	if err := service.store.Release(ctx, screeningID, seatID, userID); err != nil {
		if !errors.Is(err, ErrLockNotOwned) && !errors.Is(err, ErrLockNotFound) {
			service.recordAudit(ctx, domain.AuditLog{
				Event:       domain.AuditEventSystemError,
				UserID:      userID,
				ScreeningID: mustObjectID(screeningID),
				SeatID:      seatID,
				Message:     "release seat lock",
			})
		}
		return fmt.Errorf("release seat lock: %w", err)
	}
	service.recordAudit(ctx, domain.AuditLog{
		Event:       domain.AuditEventSeatReleased,
		UserID:      userID,
		ScreeningID: mustObjectID(screeningID),
		SeatID:      seatID,
	})

	return nil
}

func (service *Service) recordAudit(ctx context.Context, item domain.AuditLog) {
	if service.audits == nil {
		return
	}
	item.ID = bson.NewObjectID()
	item.CreatedAt = time.Now().UTC()
	if err := service.audits.Create(ctx, item); err != nil {
		log.Printf("record seat lock audit: %v", err)
	}
}

func mustObjectID(value string) bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(value)
	return id
}

func (service *Service) validateSeat(
	ctx context.Context,
	screeningID string,
	seatID string,
) (string, error) {
	id, err := bson.ObjectIDFromHex(screeningID)
	if err != nil {
		return "", ErrInvalidScreeningID
	}

	item, err := service.screenings.FindByID(ctx, id)
	if errors.Is(err, screening.ErrNotFound) {
		return "", ErrScreeningNotFound
	}
	if err != nil {
		return "", fmt.Errorf("find screening: %w", err)
	}

	seatID = strings.ToUpper(strings.TrimSpace(seatID))
	for _, seat := range item.Seats {
		if seat.ID == seatID {
			if seat.Status == domain.SeatStatusBooked {
				return "", ErrAlreadyBooked
			}
			return seatID, nil
		}
	}

	return "", ErrSeatNotFound
}
