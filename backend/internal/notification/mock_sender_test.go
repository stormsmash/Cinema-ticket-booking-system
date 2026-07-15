package notification

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/realtime"
)

func TestMockSenderWritesBookingConfirmation(t *testing.T) {
	var output bytes.Buffer
	sender := NewMockSender(log.New(&output, "", 0))
	event := realtime.SeatEvent{
		Type:        realtime.SeatBooked,
		BookingID:   "booking-1",
		ScreeningID: "screening-1",
		SeatID:      "A1",
	}

	if err := sender.SendBookingConfirmation(context.Background(), event); err != nil {
		t.Fatalf("send mock notification: %v", err)
	}
	want := "MOCK_NOTIFICATION booking_confirmed booking_id=booking-1 screening_id=screening-1 seat_id=A1"
	if !strings.Contains(output.String(), want) {
		t.Fatalf("expected %q in output, got %q", want, output.String())
	}
}

func TestMockSenderRejectsNonBookingEvent(t *testing.T) {
	var output bytes.Buffer
	sender := NewMockSender(log.New(&output, "", 0))

	err := sender.SendBookingConfirmation(context.Background(), realtime.SeatEvent{
		Type:        realtime.SeatLocked,
		ScreeningID: "screening-1",
		SeatID:      "A1",
	})
	if !errors.Is(err, ErrInvalidBookingEvent) {
		t.Fatalf("expected ErrInvalidBookingEvent, got %v", err)
	}
	if output.Len() != 0 {
		t.Fatalf("invalid event must not write a notification, got %q", output.String())
	}
}
