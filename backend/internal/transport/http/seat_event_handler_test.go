package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/realtime"
)

const seatEventTestOrigin = "http://cinema.example"

func TestSeatEventStreamForwardsMatchingEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	screeningID := bson.NewObjectID()
	hub := realtime.NewHub(10)
	defer hub.Close()

	handler := newSeatEventHandler(
		screeningServiceStub{findByID: func(context.Context, bson.ObjectID) (domain.Screening, error) {
			return domain.Screening{ID: screeningID}, nil
		}},
		hub,
		seatEventTestOrigin,
	)
	router := gin.New()
	router.GET("/screenings/:screeningID/seat-events", handler.stream)
	server := httptest.NewServer(router)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	connection, _, err := websocket.Dial(
		ctx,
		"ws"+server.URL[len("http"):]+"/screenings/"+screeningID.Hex()+"/seat-events",
		&websocket.DialOptions{HTTPHeader: http.Header{"Origin": []string{seatEventTestOrigin}}},
	)
	if err != nil {
		t.Fatalf("dial seat event stream: %v", err)
	}
	defer connection.CloseNow()

	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	want := realtime.SeatEvent{
		Version:     realtime.EventVersion,
		Type:        realtime.SeatLocked,
		ScreeningID: screeningID.Hex(),
		SeatID:      "D8",
		Status:      "LOCKED",
		ExpiresAt:   &expiresAt,
		OccurredAt:  time.Now().UTC(),
	}
	hub.Publish(want)

	var got realtime.SeatEvent
	if err := wsjson.Read(ctx, connection, &got); err != nil {
		t.Fatalf("read seat event: %v", err)
	}
	if got.Type != want.Type || got.ScreeningID != want.ScreeningID || got.SeatID != want.SeatID {
		t.Fatalf("unexpected event: %#v", got)
	}
	if got.ExpiresAt == nil || got.Status != "LOCKED" {
		t.Fatalf("expected locked event with expiry, got %#v", got)
	}
}

func TestSeatEventStreamRejectsForeignOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	screeningID := bson.NewObjectID()
	handler := newSeatEventHandler(screeningServiceStub{}, realtime.NewHub(10), seatEventTestOrigin)
	router := gin.New()
	router.GET("/screenings/:screeningID/seat-events", handler.stream)
	server := httptest.NewServer(router)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, response, err := websocket.Dial(
		ctx,
		"ws"+server.URL[len("http"):]+"/screenings/"+screeningID.Hex()+"/seat-events",
		&websocket.DialOptions{HTTPHeader: http.Header{"Origin": []string{"https://evil.example"}}},
	)
	if err == nil {
		t.Fatal("expected foreign origin handshake to fail")
	}
	if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %#v", http.StatusForbidden, response)
	}
}
