package httptransport

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/realtime"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/screening"
)

const (
	websocketPingInterval = 30 * time.Second
	websocketWriteTimeout = 5 * time.Second
)

type seatEventHandler struct {
	screenings  ScreeningService
	hub         *realtime.Hub
	frontendURL string
}

func newSeatEventHandler(
	screenings ScreeningService,
	hub *realtime.Hub,
	frontendURL string,
) *seatEventHandler {
	return &seatEventHandler{
		screenings:  screenings,
		hub:         hub,
		frontendURL: frontendURL,
	}
}

func (handler *seatEventHandler) stream(c *gin.Context) {
	if !sameOrigin(c.GetHeader("Origin"), handler.frontendURL) {
		writeError(c, http.StatusForbidden, "ORIGIN_NOT_ALLOWED", "WebSocket origin is not allowed")
		return
	}

	screeningID, err := bson.ObjectIDFromHex(c.Param("screeningID"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_SCREENING_ID", "Screening ID is invalid")
		return
	}
	if _, err := handler.screenings.FindByID(c.Request.Context(), screeningID); err != nil {
		if errors.Is(err, screening.ErrNotFound) {
			writeError(c, http.StatusNotFound, "SCREENING_NOT_FOUND", "Screening was not found")
			return
		}
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to open seat updates")
		return
	}

	events, unsubscribe, err := handler.hub.Subscribe(screeningID.Hex())
	if errors.Is(err, realtime.ErrHubFull) {
		writeError(c, http.StatusServiceUnavailable, "SEAT_EVENTS_FULL", "Too many seat update connections")
		return
	}
	if err != nil {
		writeError(c, http.StatusServiceUnavailable, "SEAT_EVENTS_UNAVAILABLE", "Seat updates are unavailable")
		return
	}
	defer unsubscribe()

	allowedOrigin, _ := url.Parse(handler.frontendURL)
	connection, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{allowedOrigin.Host},
	})
	if err != nil {
		return
	}
	defer connection.CloseNow()

	connectionContext := connection.CloseRead(context.Background())
	pingTicker := time.NewTicker(websocketPingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-connectionContext.Done():
			return
		case event, open := <-events:
			if !open {
				_ = connection.Close(websocket.StatusGoingAway, "server shutting down")
				return
			}

			writeContext, cancel := context.WithTimeout(connectionContext, websocketWriteTimeout)
			err := wsjson.Write(writeContext, connection, event)
			cancel()
			if err != nil {
				return
			}
		case <-pingTicker.C:
			pingContext, cancel := context.WithTimeout(connectionContext, websocketWriteTimeout)
			err := connection.Ping(pingContext)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func sameOrigin(origin string, frontendURL string) bool {
	if origin == "" {
		return false
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Scheme == "" || originURL.Host == "" {
		return false
	}
	allowedURL, err := url.Parse(frontendURL)
	if err != nil || allowedURL.Scheme == "" || allowedURL.Host == "" {
		return false
	}

	return strings.EqualFold(originURL.Scheme, allowedURL.Scheme) &&
		strings.EqualFold(originURL.Host, allowedURL.Host)
}
