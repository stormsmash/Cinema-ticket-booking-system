package notification

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/realtime"
)

var ErrInvalidBookingEvent = errors.New("invalid booking event")

type MockSender struct {
	logger *log.Logger
}

func NewMockSender(logger *log.Logger) *MockSender {
	return &MockSender{logger: logger}
}

func (sender *MockSender) SendBookingConfirmation(
	ctx context.Context,
	event realtime.SeatEvent,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if event.Type != realtime.SeatBooked || event.BookingID == "" ||
		event.ScreeningID == "" || event.SeatID == "" {
		return ErrInvalidBookingEvent
	}
	if sender == nil || sender.logger == nil {
		return fmt.Errorf("mock notification logger is required")
	}

	sender.logger.Printf(
		"MOCK_NOTIFICATION booking_confirmed booking_id=%s screening_id=%s seat_id=%s",
		event.BookingID,
		event.ScreeningID,
		event.SeatID,
	)
	return nil
}
