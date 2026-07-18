package httptransport

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/booking"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

type BookingService interface {
	Confirm(context.Context, string, string, string) (booking.Confirmation, error)
	ListForUser(context.Context, string) ([]booking.UserTicket, error)
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
	PriceBaht   int       `json:"price_baht"`
	TicketCode  string    `json:"ticket_code"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type userTicketResponse struct {
	ID             string    `json:"id"`
	ScreeningID    string    `json:"screening_id"`
	MovieTitle     string    `json:"movie_title"`
	AuditoriumName string    `json:"auditorium_name"`
	StartsAt       time.Time `json:"starts_at"`
	SeatID         string    `json:"seat_id"`
	PriceBaht      int       `json:"price_baht"`
	TicketCode     string    `json:"ticket_code"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
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
			PriceBaht:   item.PriceBaht,
			TicketCode:  ticketCode(item),
			Status:      string(item.Status),
			CreatedAt:   item.CreatedAt,
		},
	})
}

func (handler *bookingHandler) listMine(c *gin.Context) {
	user := authenticatedUser(c)
	tickets, err := handler.service.ListForUser(c.Request.Context(), user.ID.Hex())
	if err != nil {
		log.Printf("list user tickets: %v", err)
		writeError(c, http.StatusInternalServerError, "TICKETS_FAILED", "Unable to load tickets")
		return
	}

	data := make([]userTicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		item := ticket.Booking
		price := item.PriceBaht
		if price == 0 {
			price = ticket.Screening.TicketPriceBaht
		}
		data = append(data, userTicketResponse{
			ID:             item.ID.Hex(),
			ScreeningID:    item.ScreeningID.Hex(),
			MovieTitle:     ticket.Screening.Movie.Title,
			AuditoriumName: ticket.Screening.Auditorium.Name,
			StartsAt:       ticket.Screening.StartsAt,
			SeatID:         item.SeatID,
			PriceBaht:      price,
			TicketCode:     ticketCode(item),
			Status:         string(item.Status),
			CreatedAt:      item.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func ticketCode(item domain.Booking) string {
	return "LUMINA-" + item.ID.Hex()
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
