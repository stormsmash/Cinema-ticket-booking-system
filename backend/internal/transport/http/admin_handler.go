package httptransport

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/admin"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

const (
	defaultAdminPageSize = int64(20)
	maximumAdminPageSize = int64(100)
	maximumAdminPage     = int64(10_000)
)

var errInvalidAdminFilters = errors.New("invalid admin filters")

type AdminService interface {
	ListBookings(context.Context, admin.BookingFilter) (admin.BookingPage, error)
	ListAuditLogs(context.Context, admin.AuditFilter) (admin.AuditPage, error)
}

type adminHandler struct {
	service AdminService
}

type adminBookingResponse struct {
	ID        string                   `json:"id"`
	SeatID    string                   `json:"seat_id"`
	Status    string                   `json:"status"`
	CreatedAt time.Time                `json:"created_at"`
	User      adminBookingUserResponse `json:"user"`
	Screening adminScreeningResponse   `json:"screening"`
}

type adminBookingUserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type adminScreeningResponse struct {
	ID             string    `json:"id"`
	MovieTitle     string    `json:"movie_title"`
	AuditoriumName string    `json:"auditorium_name"`
	StartsAt       time.Time `json:"starts_at"`
}

type adminAuditResponse struct {
	ID          string    `json:"id"`
	Event       string    `json:"event"`
	BookingID   string    `json:"booking_id,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	ScreeningID string    `json:"screening_id,omitempty"`
	SeatID      string    `json:"seat_id,omitempty"`
	Message     string    `json:"message,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type pageMetaResponse struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func newAdminHandler(service AdminService) *adminHandler {
	return &adminHandler{service: service}
}

func (handler *adminHandler) bookings(c *gin.Context) {
	filter, err := parseBookingFilters(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_ADMIN_FILTERS", "Admin filters are invalid")
		return
	}

	page, err := handler.service.ListBookings(c.Request.Context(), filter)
	if err != nil {
		log.Printf("list admin bookings: %v", err)
		writeError(c, http.StatusInternalServerError, "ADMIN_QUERY_FAILED", "Unable to load bookings")
		return
	}

	items := make([]adminBookingResponse, 0, len(page.Items))
	for _, item := range page.Items {
		userID := item.User.ID.Hex()
		if item.User.ID.IsZero() {
			userID = item.Booking.UserID
		}
		items = append(items, adminBookingResponse{
			ID:        item.Booking.ID.Hex(),
			SeatID:    item.Booking.SeatID,
			Status:    string(item.Booking.Status),
			CreatedAt: item.Booking.CreatedAt,
			User: adminBookingUserResponse{
				ID:    userID,
				Name:  item.User.Name,
				Email: item.User.Email,
			},
			Screening: adminScreeningResponse{
				ID:             item.Booking.ScreeningID.Hex(),
				MovieTitle:     item.Screening.Movie.Title,
				AuditoriumName: item.Screening.Auditorium.Name,
				StartsAt:       item.Screening.StartsAt,
			},
		})
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": pageMeta(filter.Page, filter.PageSize, page.Total),
	})
}

func (handler *adminHandler) auditLogs(c *gin.Context) {
	filter, err := parseAuditFilters(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_ADMIN_FILTERS", "Admin filters are invalid")
		return
	}

	page, err := handler.service.ListAuditLogs(c.Request.Context(), filter)
	if err != nil {
		log.Printf("list admin audit logs: %v", err)
		writeError(c, http.StatusInternalServerError, "ADMIN_QUERY_FAILED", "Unable to load audit logs")
		return
	}

	items := make([]adminAuditResponse, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, adminAuditResponse{
			ID:          item.ID.Hex(),
			Event:       string(item.Event),
			BookingID:   objectIDString(item.BookingID),
			UserID:      item.UserID,
			ScreeningID: objectIDString(item.ScreeningID),
			SeatID:      item.SeatID,
			Message:     item.Message,
			CreatedAt:   item.CreatedAt,
		})
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"meta": pageMeta(filter.Page, filter.PageSize, page.Total),
	})
}

func parseBookingFilters(c *gin.Context) (admin.BookingFilter, error) {
	page, pageSize, err := parseAdminPagination(c)
	if err != nil {
		return admin.BookingFilter{}, err
	}
	movie := strings.TrimSpace(c.Query("movie"))
	if len(movie) > 120 {
		return admin.BookingFilter{}, errInvalidAdminFilters
	}
	status := domain.BookingStatus(strings.ToUpper(strings.TrimSpace(c.Query("status"))))
	if status != "" && status != domain.BookingStatusHolding && status != domain.BookingStatusBooked &&
		status != domain.BookingStatusTimedOut && status != domain.BookingStatusCancelled {
		return admin.BookingFilter{}, errInvalidAdminFilters
	}

	return admin.BookingFilter{Movie: movie, Status: status, Page: page, PageSize: pageSize}, nil
}

func parseAuditFilters(c *gin.Context) (admin.AuditFilter, error) {
	page, pageSize, err := parseAdminPagination(c)
	if err != nil {
		return admin.AuditFilter{}, err
	}
	event := domain.AuditEvent(strings.ToUpper(strings.TrimSpace(c.Query("event"))))
	if event != "" && event != domain.AuditEventBookingSuccess &&
		event != domain.AuditEventBookingTimeout && event != domain.AuditEventSeatReleased &&
		event != domain.AuditEventSystemError {
		return admin.AuditFilter{}, errInvalidAdminFilters
	}

	return admin.AuditFilter{Event: event, Page: page, PageSize: pageSize}, nil
}

func parseAdminPagination(c *gin.Context) (int64, int64, error) {
	page, err := positiveQueryInteger(c.Query("page"), 1)
	if err != nil || page > maximumAdminPage {
		return 0, 0, errInvalidAdminFilters
	}
	pageSize, err := positiveQueryInteger(c.Query("page_size"), defaultAdminPageSize)
	if err != nil || pageSize > maximumAdminPageSize {
		return 0, 0, errInvalidAdminFilters
	}
	return page, pageSize, nil
}

func positiveQueryInteger(value string, fallback int64) (int64, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, errInvalidAdminFilters
	}
	return parsed, nil
}

func pageMeta(page, pageSize, total int64) pageMetaResponse {
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return pageMetaResponse{Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}
}

func objectIDString(id bson.ObjectID) string {
	if id.IsZero() {
		return ""
	}
	return id.Hex()
}
