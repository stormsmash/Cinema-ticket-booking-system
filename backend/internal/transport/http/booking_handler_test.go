package httptransport

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/booking"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/health"
)

type bookingServiceStub struct {
	confirmation booking.Confirmation
	err          error
	userID       string
	tickets      []booking.UserTicket
	listError    error
}

func (stub *bookingServiceStub) ListForUser(
	_ context.Context,
	userID string,
) ([]booking.UserTicket, error) {
	stub.userID = userID
	return stub.tickets, stub.listError
}

func (stub *bookingServiceStub) Confirm(
	_ context.Context,
	_ string,
	_ string,
	userID string,
) (booking.Confirmation, error) {
	stub.userID = userID
	return stub.confirmation, stub.err
}

func bookingTestRouter(user domain.User, bookings BookingService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(Dependencies{
		Readiness: readinessFunc(func(context.Context) health.Report {
			return health.Report{Status: health.StatusReady}
		}),
		Screenings: screeningServiceStub{},
		Auth:       &authServiceStub{user: user},
		AuthConfig: AuthHandlerConfig{FrontendURL: "http://localhost:3000", SessionTTL: time.Hour},
		Bookings:   bookings,
	})
}

func TestConfirmBookingRequiresAuthentication(t *testing.T) {
	router := bookingTestRouter(domain.User{}, &bookingServiceStub{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/bookings",
		bytes.NewBufferString(`{"screening_id":"66a000000000000000000001","seat_id":"A1"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestConfirmBookingUsesAuthenticatedUser(t *testing.T) {
	user := domain.User{ID: bson.NewObjectID()}
	bookingID := bson.NewObjectID()
	screeningID := bson.NewObjectID()
	service := &bookingServiceStub{confirmation: booking.Confirmation{
		Created: true,
		Booking: domain.Booking{
			ID:          bookingID,
			ScreeningID: screeningID,
			SeatID:      "A1",
			Status:      domain.BookingStatusBooked,
			CreatedAt:   time.Now().UTC(),
		},
	}}
	router := bookingTestRouter(user, service)
	body := `{"screening_id":"` + screeningID.Hex() + `","seat_id":"A1","user_id":"forged"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body)
	}
	if service.userID != user.ID.Hex() {
		t.Fatalf("expected authenticated user %q, got %q", user.ID.Hex(), service.userID)
	}
	if strings.Contains(recorder.Body.String(), "forged") ||
		!strings.Contains(recorder.Body.String(), bookingID.Hex()) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestConfirmBookingMapsExpiredLock(t *testing.T) {
	user := domain.User{ID: bson.NewObjectID()}
	router := bookingTestRouter(user, &bookingServiceStub{err: booking.ErrSeatLockExpired})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/bookings",
		bytes.NewBufferString(`{"screening_id":"66a000000000000000000001","seat_id":"A1"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict ||
		!strings.Contains(recorder.Body.String(), "SEAT_LOCK_EXPIRED") {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestListMyTicketsUsesAuthenticatedUser(t *testing.T) {
	user := domain.User{ID: bson.NewObjectID()}
	bookingID := bson.NewObjectID()
	screeningID := bson.NewObjectID()
	service := &bookingServiceStub{tickets: []booking.UserTicket{{
		Booking: domain.Booking{
			ID:          bookingID,
			ScreeningID: screeningID,
			SeatID:      "C5",
			PriceBaht:   240,
			Status:      domain.BookingStatusBooked,
			CreatedAt:   time.Now().UTC(),
		},
		Screening: domain.Screening{
			ID:              screeningID,
			Movie:           domain.Movie{Title: "ธี่หยด 2"},
			Auditorium:      domain.Auditorium{Name: "โรง 2"},
			StartsAt:        time.Now().UTC().Add(time.Hour),
			TicketPriceBaht: 240,
		},
	}}}
	router := bookingTestRouter(user, service)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/bookings/me", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || service.userID != user.ID.Hex() {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "ธี่หยด 2") ||
		!strings.Contains(recorder.Body.String(), "TICKET-"+bookingID.Hex()) {
		t.Fatalf("ticket details missing: %s", recorder.Body.String())
	}
}
