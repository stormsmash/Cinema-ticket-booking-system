package httptransport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/health"
)

type readinessFunc func(context.Context) health.Report

func (function readinessFunc) Check(ctx context.Context) health.Report {
	return function(ctx)
}

func TestLiveness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(readinessFunc(func(context.Context) health.Report {
		return health.Report{Status: health.StatusReady}
	}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Fatalf("expected health status ok, got %q", response.Status)
	}
}

func TestReadinessReturnsServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(readinessFunc(func(context.Context) health.Report {
		return health.Report{
			Status: health.StatusNotReady,
			Checks: map[string]string{"redis": health.CheckFailed},
		}
	}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
}
