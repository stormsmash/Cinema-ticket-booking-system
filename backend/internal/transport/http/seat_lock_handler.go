package httptransport

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/seatlock"
)

type SeatLockService interface {
	Acquire(context.Context, string, string, string) (seatlock.Lock, error)
	CurrentLocks(context.Context, string, []domain.Seat) (map[string]seatlock.Lock, error)
	Release(context.Context, string, string, string) error
}

type seatLockHandler struct {
	service SeatLockService
}

type seatLockResponse struct {
	ScreeningID string    `json:"screening_id"`
	SeatID      string    `json:"seat_id"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func newSeatLockHandler(service SeatLockService) *seatLockHandler {
	return &seatLockHandler{service: service}
}

func (handler *seatLockHandler) acquire(c *gin.Context) {
	user := authenticatedUser(c)
	lock, err := handler.service.Acquire(
		c.Request.Context(),
		c.Param("screeningID"),
		c.Param("seatID"),
		user.ID.Hex(),
	)
	if err != nil {
		handler.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": seatLockResponse{
			ScreeningID: lock.ScreeningID,
			SeatID:      lock.SeatID,
			Status:      string(domain.SeatStatusLocked),
			ExpiresAt:   lock.ExpiresAt,
		},
	})
}

func (handler *seatLockHandler) release(c *gin.Context) {
	user := authenticatedUser(c)
	err := handler.service.Release(
		c.Request.Context(),
		c.Param("screeningID"),
		c.Param("seatID"),
		user.ID.Hex(),
	)
	if err != nil {
		handler.writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (handler *seatLockHandler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, seatlock.ErrInvalidScreeningID):
		writeError(c, http.StatusBadRequest, "INVALID_SCREENING_ID", "Screening ID is invalid")
	case errors.Is(err, seatlock.ErrScreeningNotFound):
		writeError(c, http.StatusNotFound, "SCREENING_NOT_FOUND", "Screening was not found")
	case errors.Is(err, seatlock.ErrSeatNotFound):
		writeError(c, http.StatusNotFound, "SEAT_NOT_FOUND", "Seat was not found")
	case errors.Is(err, seatlock.ErrAlreadyLocked):
		writeError(c, http.StatusConflict, "SEAT_ALREADY_LOCKED", "Seat is locked by another user")
	case errors.Is(err, seatlock.ErrAlreadyBooked):
		writeError(c, http.StatusConflict, "SEAT_ALREADY_BOOKED", "Seat is already booked")
	case errors.Is(err, seatlock.ErrLockNotOwned):
		writeError(c, http.StatusConflict, "SEAT_LOCK_NOT_OWNED", "Seat is locked by another user")
	default:
		log.Printf("seat lock operation: %v", err)
		writeError(c, http.StatusInternalServerError, "SEAT_LOCK_FAILED", "Unable to update seat lock")
	}
}
