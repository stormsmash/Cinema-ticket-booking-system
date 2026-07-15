package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/admin"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/health"
)

type adminServiceStub struct {
	bookingPage   admin.BookingPage
	auditPage     admin.AuditPage
	bookingFilter admin.BookingFilter
	auditFilter   admin.AuditFilter
	bookingCalls  int
	auditCalls    int
}

func (stub *adminServiceStub) ListBookings(
	_ context.Context,
	filter admin.BookingFilter,
) (admin.BookingPage, error) {
	stub.bookingCalls++
	stub.bookingFilter = filter
	return stub.bookingPage, nil
}

func (stub *adminServiceStub) ListAuditLogs(
	_ context.Context,
	filter admin.AuditFilter,
) (admin.AuditPage, error) {
	stub.auditCalls++
	stub.auditFilter = filter
	return stub.auditPage, nil
}

func adminTestRouter(user domain.User, service AdminService) http.Handler {
	return NewRouter(Dependencies{
		Readiness: readinessFunc(func(context.Context) health.Report {
			return health.Report{Status: health.StatusReady}
		}),
		Screenings: screeningServiceStub{},
		Auth:       &authServiceStub{user: user},
		Admin:      service,
		AuthConfig: AuthHandlerConfig{FrontendURL: "http://localhost:3000", SessionTTL: time.Hour},
	})
}

func TestAdminBookingsRequiresAuthentication(t *testing.T) {
	service := &adminServiceStub{}
	router := adminTestRouter(domain.User{}, service)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/bookings", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized || service.bookingCalls != 0 {
		t.Fatalf("expected 401 without service call, status=%d calls=%d", recorder.Code, service.bookingCalls)
	}
}

func TestAdminBookingsRejectsNonAdminRoles(t *testing.T) {
	for _, role := range []domain.UserRole{"", domain.UserRoleUser, "UNKNOWN"} {
		t.Run(string(role), func(t *testing.T) {
			service := &adminServiceStub{}
			router := adminTestRouter(domain.User{ID: bson.NewObjectID(), Role: role}, service)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/bookings", nil)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusForbidden || service.bookingCalls != 0 {
				t.Fatalf("expected 403 without service call, status=%d calls=%d", recorder.Code, service.bookingCalls)
			}
		})
	}
}

func TestAdminBookingsParsesTypedFilters(t *testing.T) {
	bookingID := bson.NewObjectID()
	userID := bson.NewObjectID()
	screeningID := bson.NewObjectID()
	service := &adminServiceStub{bookingPage: admin.BookingPage{
		Items: []admin.BookingItem{{
			Booking: domain.Booking{
				ID:          bookingID,
				UserID:      userID.Hex(),
				ScreeningID: screeningID,
				SeatID:      "A1",
				Status:      domain.BookingStatusBooked,
				CreatedAt:   time.Now().UTC(),
			},
			User: domain.User{ID: userID, Name: "Viewer", Email: "viewer@example.com"},
			Screening: domain.Screening{
				ID:         screeningID,
				Movie:      domain.Movie{Title: "Midnight Signal"},
				Auditorium: domain.Auditorium{Name: "Hall 1"},
			},
		}},
		Total: 1,
	}}
	router := adminTestRouter(domain.User{ID: bson.NewObjectID(), Role: domain.UserRoleAdmin}, service)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/bookings?movie=Midnight%20Signal&status=booked&page=2&page_size=5",
		nil,
	)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || service.bookingCalls != 1 {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if service.bookingFilter.Movie != "Midnight Signal" ||
		service.bookingFilter.Status != domain.BookingStatusBooked ||
		service.bookingFilter.Page != 2 || service.bookingFilter.PageSize != 5 {
		t.Fatalf("unexpected filter: %#v", service.bookingFilter)
	}
	if recorder.Header().Get("Cache-Control") != "no-store" ||
		strings.Contains(recorder.Body.String(), "google_subject") {
		t.Fatalf("unexpected admin response headers/body: %s", recorder.Body.String())
	}
}

func TestAdminBookingsRejectsInvalidPaginationAndStatus(t *testing.T) {
	for _, query := range []string{
		"status=%24ne",
		"page=0",
		"page=10001",
		"page_size=101",
		"page=not-a-number",
	} {
		t.Run(query, func(t *testing.T) {
			service := &adminServiceStub{}
			router := adminTestRouter(domain.User{ID: bson.NewObjectID(), Role: domain.UserRoleAdmin}, service)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/bookings?"+query, nil)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest || service.bookingCalls != 0 {
				t.Fatalf("expected 400 without service call, status=%d calls=%d", recorder.Code, service.bookingCalls)
			}
		})
	}
}

func TestAdminAuditLogsAreReadOnlyAndFiltered(t *testing.T) {
	service := &adminServiceStub{auditPage: admin.AuditPage{Items: []domain.AuditLog{{
		ID:        bson.NewObjectID(),
		Event:     domain.AuditEventSeatReleased,
		SeatID:    "A1",
		CreatedAt: time.Now().UTC(),
	}}, Total: 1}}
	router := adminTestRouter(domain.User{ID: bson.NewObjectID(), Role: domain.UserRoleAdmin}, service)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/admin/audit-logs?event=seat_released",
		nil,
	)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || service.auditCalls != 1 ||
		service.auditFilter.Event != domain.AuditEventSeatReleased {
		t.Fatalf("unexpected response/filter: status=%d filter=%#v", recorder.Code, service.auditFilter)
	}
}
