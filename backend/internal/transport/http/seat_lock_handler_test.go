package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/health"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/seatlock"
)

type seatLockServiceStub struct {
	lock         seatlock.Lock
	acquireError error
	releaseError error
	currentLocks map[string]seatlock.Lock
}

func (stub *seatLockServiceStub) Acquire(
	context.Context,
	string,
	string,
	string,
) (seatlock.Lock, error) {
	return stub.lock, stub.acquireError
}

func (stub *seatLockServiceStub) CurrentLocks(
	context.Context,
	string,
	[]domain.Seat,
) (map[string]seatlock.Lock, error) {
	return stub.currentLocks, nil
}

func (stub *seatLockServiceStub) Release(context.Context, string, string, string) error {
	return stub.releaseError
}

func seatLockTestRouter(
	user domain.User,
	locks SeatLockService,
	screenings ScreeningService,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(Dependencies{
		Readiness: readinessFunc(func(context.Context) health.Report {
			return health.Report{Status: health.StatusReady}
		}),
		Screenings: screenings,
		Auth:       &authServiceStub{user: user},
		AuthConfig: AuthHandlerConfig{FrontendURL: "http://localhost:3000", SessionTTL: time.Hour},
		SeatLocks:  locks,
	})
}

func TestAcquireSeatLockRequiresAuthentication(t *testing.T) {
	router := seatLockTestRouter(domain.User{}, &seatLockServiceStub{}, screeningServiceStub{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/screenings/66a000000000000000000001/seats/A1/lock",
		nil,
	)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestAcquireSeatLockReturnsExpiry(t *testing.T) {
	user := domain.User{ID: bson.NewObjectID()}
	lock := seatlock.Lock{
		ScreeningID: "66a000000000000000000001",
		SeatID:      "A1",
		UserID:      user.ID.Hex(),
		ExpiresAt:   time.Now().UTC().Add(10 * time.Minute),
	}
	router := seatLockTestRouter(user, &seatLockServiceStub{lock: lock}, screeningServiceStub{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/screenings/66a000000000000000000001/seats/A1/lock",
		nil,
	)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK ||
		!strings.Contains(recorder.Body.String(), `"status":"LOCKED"`) ||
		!strings.Contains(recorder.Body.String(), `"seat_id":"A1"`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAcquireSeatLockReturnsConflict(t *testing.T) {
	user := domain.User{ID: bson.NewObjectID()}
	router := seatLockTestRouter(
		user,
		&seatLockServiceStub{acquireError: seatlock.ErrAlreadyLocked},
		screeningServiceStub{},
	)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/screenings/66a000000000000000000001/seats/A1/lock",
		nil,
	)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict ||
		!strings.Contains(recorder.Body.String(), "SEAT_ALREADY_LOCKED") {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestReleaseSeatLockRejectsDifferentOwner(t *testing.T) {
	user := domain.User{ID: bson.NewObjectID()}
	router := seatLockTestRouter(
		user,
		&seatLockServiceStub{releaseError: seatlock.ErrLockNotOwned},
		screeningServiceStub{},
	)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/screenings/66a000000000000000000001/seats/A1/lock",
		nil,
	)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict ||
		!strings.Contains(recorder.Body.String(), "SEAT_LOCK_NOT_OWNED") {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSeatMapMarksCurrentUsersLock(t *testing.T) {
	user := domain.User{ID: bson.NewObjectID()}
	screeningID := bson.NewObjectID()
	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	locks := &seatLockServiceStub{currentLocks: map[string]seatlock.Lock{
		"A1": {
			ScreeningID: screeningID.Hex(),
			SeatID:      "A1",
			UserID:      user.ID.Hex(),
			ExpiresAt:   expiresAt,
		},
	}}
	screenings := screeningServiceStub{findByID: func(context.Context, bson.ObjectID) (domain.Screening, error) {
		return domain.Screening{
			ID:    screeningID,
			Seats: []domain.Seat{{ID: "A1", Row: "A", Number: 1}},
		}, nil
	}}
	router := seatLockTestRouter(user, locks, screenings)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/screenings/"+screeningID.Hex()+"/seats",
		nil,
	)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK ||
		!strings.Contains(recorder.Body.String(), `"status":"LOCKED"`) ||
		!strings.Contains(recorder.Body.String(), `"locked_by_me":true`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSeatMapKeepsBookedStatusAheadOfStaleLock(t *testing.T) {
	screeningID := bson.NewObjectID()
	locks := &seatLockServiceStub{currentLocks: map[string]seatlock.Lock{
		"A1": {ScreeningID: screeningID.Hex(), SeatID: "A1", UserID: "user-1"},
	}}
	screenings := screeningServiceStub{findByID: func(context.Context, bson.ObjectID) (domain.Screening, error) {
		return domain.Screening{
			ID: screeningID,
			Seats: []domain.Seat{{
				ID:     "A1",
				Row:    "A",
				Number: 1,
				Status: domain.SeatStatusBooked,
			}},
		}, nil
	}}
	router := seatLockTestRouter(domain.User{}, locks, screenings)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/screenings/"+screeningID.Hex()+"/seats",
		nil,
	)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK ||
		!strings.Contains(recorder.Body.String(), `"status":"BOOKED"`) ||
		strings.Contains(recorder.Body.String(), `"locked_by_me":true`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
