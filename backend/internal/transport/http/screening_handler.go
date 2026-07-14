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
	service ScreeningService
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
	ID     string            `json:"id"`
	Row    string            `json:"row"`
	Number int               `json:"number"`
	Status domain.SeatStatus `json:"status"`
}

type seatMapData struct {
	ScreeningID string             `json:"screening_id"`
	Movie       movieResponse      `json:"movie"`
	Auditorium  auditoriumResponse `json:"auditorium"`
	StartsAt    time.Time          `json:"starts_at"`
	Seats       []seatResponse     `json:"seats"`
}

func newScreeningHandler(service ScreeningService) *screeningHandler {
	return &screeningHandler{service: service}
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

	seats := make([]seatResponse, 0, len(item.Seats))
	for _, seat := range item.Seats {
		seats = append(seats, seatResponse{
			ID:     seat.ID,
			Row:    seat.Row,
			Number: seat.Number,
			Status: domain.SeatStatusAvailable,
		})
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
