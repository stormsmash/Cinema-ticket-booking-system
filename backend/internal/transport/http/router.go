package httptransport

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/health"
)

type ReadinessChecker interface {
	Check(context.Context) health.Report
}

type Dependencies struct {
	Readiness  ReadinessChecker
	Screenings ScreeningService
}

func NewRouter(dependencies Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	api := router.Group("/api/v1")
	api.GET("/health", liveness)
	api.GET("/health/live", liveness)
	api.GET("/health/ready", ready(dependencies.Readiness))

	screenings := newScreeningHandler(dependencies.Screenings)
	api.GET("/screenings", screenings.list)
	api.GET("/screenings/:screeningID/seats", screenings.seats)

	return router
}

func liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func ready(readiness ReadinessChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		report := readiness.Check(ctx)
		statusCode := http.StatusOK
		if report.Status != health.StatusReady {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, report)
	}
}
