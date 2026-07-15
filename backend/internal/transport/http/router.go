package httptransport

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/health"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/realtime"
)

type ReadinessChecker interface {
	Check(context.Context) health.Report
}

type Dependencies struct {
	Readiness   ReadinessChecker
	Screenings  ScreeningService
	Auth        AuthService
	AuthConfig  AuthHandlerConfig
	SeatLocks   SeatLockService
	SeatEvents  *realtime.Hub
	FrontendURL string
	Bookings    BookingService
	Admin       AdminService
}

func NewRouter(dependencies Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{SkipQueryString: true}),
		gin.Recovery(),
	)

	api := router.Group("/api/v1")
	api.GET("/health", liveness)
	api.GET("/health/live", liveness)
	api.GET("/health/ready", ready(dependencies.Readiness))

	auth := newAuthHandler(dependencies.Auth, dependencies.AuthConfig)
	api.GET("/auth/config", auth.configuration)
	api.GET("/auth/google", auth.google)
	api.GET("/auth/google/callback", auth.googleCallback)
	api.GET("/auth/me", auth.requireAuthentication(), auth.me)
	api.POST("/auth/logout", auth.logout)

	screenings := newScreeningHandler(dependencies.Screenings, dependencies.SeatLocks)
	seatLocks := newSeatLockHandler(dependencies.SeatLocks)
	seatEvents := newSeatEventHandler(
		dependencies.Screenings,
		dependencies.SeatEvents,
		dependencies.FrontendURL,
	)
	bookings := newBookingHandler(dependencies.Bookings)
	adminHandler := newAdminHandler(dependencies.Admin)
	api.GET("/screenings", screenings.list)
	api.GET(
		"/screenings/:screeningID/seats",
		auth.optionalAuthentication(),
		screenings.seats,
	)
	api.POST(
		"/screenings/:screeningID/seats/:seatID/lock",
		auth.requireAuthentication(),
		seatLocks.acquire,
	)
	adminRoutes := api.Group(
		"/admin",
		auth.requireAuthentication(),
		auth.requireAdmin(),
	)
	adminRoutes.GET("/bookings", adminHandler.bookings)
	adminRoutes.GET("/audit-logs", adminHandler.auditLogs)
	api.DELETE(
		"/screenings/:screeningID/seats/:seatID/lock",
		auth.requireAuthentication(),
		seatLocks.release,
	)
	api.GET(
		"/screenings/:screeningID/seat-events",
		seatEvents.stream,
	)
	api.POST(
		"/bookings",
		auth.requireAuthentication(),
		bookings.confirm,
	)

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
