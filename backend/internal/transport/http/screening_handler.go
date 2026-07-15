package httptransport

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/screening"
)

type ScreeningService interface {
	List(context.Context) ([]domain.Screening, error)
	FindByID(context.Context, bson.ObjectID) (domain.Screening, error)
}

type screeningHandler struct {
	service   ScreeningService
	seatLocks SeatLockService
}

type movieResponse struct {
	Title           string `json:"title"`
	DurationMinutes int    `json:"duration_minutes"`
}

type auditoriumResponse struct {
	Name        string `json:"name"`
	Rows        int    `json:"rows"`
	SeatsPerRow int    `json:"seats_per_row"`
}

type screeningSummary struct {
	ID         string             `json:"id"`
	Movie      movieResponse      `json:"movie"`
	Auditorium auditoriumResponse `json:"auditorium"`
	StartsAt   time.Time          `json:"starts_at"`
}

type seatResponse struct {
	ID            string            `json:"id"`
	Row           string            `json:"row"`
	Number        int               `json:"number"`
	Status        domain.SeatStatus `json:"status"`
	LockedByMe    bool              `json:"locked_by_me"`
	LockExpiresAt *time.Time        `json:"lock_expires_at,omitempty"`
}

type seatMapData struct {
	ScreeningID string             `json:"screening_id"`
	Movie       movieResponse      `json:"movie"`
	Auditorium  auditoriumResponse `json:"auditorium"`
	StartsAt    time.Time          `json:"starts_at"`
	Seats       []seatResponse     `json:"seats"`
}

func newScreeningHandler(service ScreeningService, seatLocks SeatLockService) *screeningHandler {
	return &screeningHandler{service: service, seatLocks: seatLocks}
}

func (handler *screeningHandler) list(c *gin.Context) {
	screenings, err := handler.service.List(c.Request.Context())
	if err != nil {
		log.Printf("list screenings: %v", err)
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to list screenings")
		return
	}

	data := make([]screeningSummary, 0, len(screenings))
	for _, item := range screenings {
		data = append(data, toScreeningSummary(item))
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (handler *screeningHandler) seats(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("screeningID"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_SCREENING_ID", "Screening ID is invalid")
		return
	}

	item, err := handler.service.FindByID(c.Request.Context(), id)
	if errors.Is(err, screening.ErrNotFound) {
		writeError(c, http.StatusNotFound, "SCREENING_NOT_FOUND", "Screening was not found")
		return
	}
	if err != nil {
		log.Printf("get screening seats: %v", err)
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load the seat map")
		return
	}

	locks, err := handler.seatLocks.CurrentLocks(c.Request.Context(), item.ID.Hex(), item.Seats)
	if err != nil {
		log.Printf("get screening seat locks: %v", err)
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load the seat map")
		return
	}

	currentUserID := ""
	if user, ok := optionalUser(c); ok {
		currentUserID = user.ID.Hex()
	}

	seats := make([]seatResponse, 0, len(item.Seats))
	for _, seat := range item.Seats {
		status := seat.Status
		if status == "" {
			status = domain.SeatStatusAvailable
		}
		response := seatResponse{
			ID:     seat.ID,
			Row:    seat.Row,
			Number: seat.Number,
			Status: status,
		}
		if lock, exists := locks[seat.ID]; exists && status == domain.SeatStatusAvailable {
			expiresAt := lock.ExpiresAt
			response.Status = domain.SeatStatusLocked
			response.LockedByMe = lock.UserID == currentUserID
			response.LockExpiresAt = &expiresAt
		}

		seats = append(seats, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": seatMapData{
			ScreeningID: item.ID.Hex(),
			Movie:       toMovieResponse(item.Movie),
			Auditorium:  toAuditoriumResponse(item.Auditorium),
			StartsAt:    item.StartsAt,
			Seats:       seats,
		},
	})
}

func toScreeningSummary(item domain.Screening) screeningSummary {
	return screeningSummary{
		ID:         item.ID.Hex(),
		Movie:      toMovieResponse(item.Movie),
		Auditorium: toAuditoriumResponse(item.Auditorium),
		StartsAt:   item.StartsAt,
	}
}

func toMovieResponse(movie domain.Movie) movieResponse {
	return movieResponse{
		Title:           movie.Title,
		DurationMinutes: movie.DurationMinutes,
	}
}

func toAuditoriumResponse(auditorium domain.Auditorium) auditoriumResponse {
	return auditoriumResponse{
		Name:        auditorium.Name,
		Rows:        auditorium.Rows,
		SeatsPerRow: auditorium.SeatsPerRow,
	}
}
