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
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/screening"
)

func TestListScreenings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := bson.NewObjectID()
	router := NewRouter(Dependencies{
		Readiness: readinessFunc(func(context.Context) health.Report {
			return health.Report{Status: health.StatusReady}
		}),
		Screenings: screeningServiceStub{
			list: func(context.Context) ([]domain.Screening, error) {
				return []domain.Screening{
					{
						ID:       id,
						Movie:    domain.Movie{Title: "Midnight Signal", DurationMinutes: 118},
						StartsAt: time.Date(2026, time.July, 15, 19, 0, 0, 0, time.UTC),
					},
				}, nil
			},
		},
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/screenings", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), id.Hex()) ||
		!strings.Contains(recorder.Body.String(), "Midnight Signal") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestSeatMapRejectsInvalidScreeningID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{
		Readiness:  readinessFunc(func(context.Context) health.Report { return health.Report{} }),
		Screenings: screeningServiceStub{},
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/screenings/not-an-id/seats", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "INVALID_SCREENING_ID") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestSeatMapReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := bson.NewObjectID()
	router := NewRouter(Dependencies{
		Readiness: readinessFunc(func(context.Context) health.Report { return health.Report{} }),
		Screenings: screeningServiceStub{
			findByID: func(context.Context, bson.ObjectID) (domain.Screening, error) {
				return domain.Screening{}, screening.ErrNotFound
			},
		},
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/screenings/"+id.Hex()+"/seats",
		nil,
	)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}
