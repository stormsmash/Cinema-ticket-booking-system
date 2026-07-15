package httptransport

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/booking"
)

type BookingService interface {
	Confirm(context.Context, string, string, string) (booking.Confirmation, error)
}

type bookingHandler struct {
	service BookingService
}

type confirmBookingRequest struct {
	ScreeningID string `json:"screening_id" binding:"required"`
	SeatID      string `json:"seat_id" binding:"required"`
}

type bookingResponse struct {
	ID          string    `json:"id"`
	ScreeningID string    `json:"screening_id"`
	SeatID      string    `json:"seat_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func newBookingHandler(service BookingService) *bookingHandler {
	return &bookingHandler{service: service}
}

func (handler *bookingHandler) confirm(c *gin.Context) {
	var request confirmBookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_BOOKING_REQUEST", "Screening and seat are required")
		return
	}

	user := authenticatedUser(c)
	confirmation, err := handler.service.Confirm(
		c.Request.Context(),
		request.ScreeningID,
		request.SeatID,
		user.ID.Hex(),
	)
	if err != nil {
		handler.writeServiceError(c, err)
		return
	}

	statusCode := http.StatusCreated
	if !confirmation.Created {
		statusCode = http.StatusOK
	}
	item := confirmation.Booking
	c.JSON(statusCode, gin.H{
		"data": bookingResponse{
			ID:          item.ID.Hex(),
			ScreeningID: item.ScreeningID.Hex(),
			SeatID:      item.SeatID,
			Status:      string(item.Status),
			CreatedAt:   item.CreatedAt,
		},
	})
}

func (handler *bookingHandler) writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, booking.ErrInvalidScreeningID):
		writeError(c, http.StatusBadRequest, "INVALID_SCREENING_ID", "Screening ID is invalid")
	case errors.Is(err, booking.ErrScreeningNotFound):
		writeError(c, http.StatusNotFound, "SCREENING_NOT_FOUND", "Screening was not found")
	case errors.Is(err, booking.ErrSeatNotFound):
		writeError(c, http.StatusNotFound, "SEAT_NOT_FOUND", "Seat was not found")
	case errors.Is(err, booking.ErrScreeningStarted):
		writeError(c, http.StatusConflict, "SCREENING_STARTED", "This screening has already started")
	case errors.Is(err, booking.ErrSeatLockExpired):
		writeError(c, http.StatusConflict, "SEAT_LOCK_EXPIRED", "The seat hold has expired")
	case errors.Is(err, booking.ErrSeatLockNotOwned):
		writeError(c, http.StatusConflict, "SEAT_LOCK_NOT_OWNED", "The seat is held by another user")
	case errors.Is(err, booking.ErrSeatAlreadyBooked):
		writeError(c, http.StatusConflict, "SEAT_ALREADY_BOOKED", "The seat is already booked")
	default:
		log.Printf("confirm booking: %v", err)
		writeError(c, http.StatusInternalServerError, "BOOKING_FAILED", "Unable to confirm the booking")
	}
}
