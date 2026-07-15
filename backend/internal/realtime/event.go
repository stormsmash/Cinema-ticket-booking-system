package realtime

import "time"

const EventVersion = 1

type SeatEventType string

const (
	SeatLocked   SeatEventType = "seat.locked"
	SeatReleased SeatEventType = "seat.released"
	SeatExpired  SeatEventType = "seat.expired"
	SeatBooked   SeatEventType = "seat.booked"
)

type SeatEvent struct {
	Version     int           `json:"version"`
	Type        SeatEventType `json:"type"`
	ScreeningID string        `json:"screening_id"`
	SeatID      string        `json:"seat_id"`
	Status      string        `json:"status"`
	ExpiresAt   *time.Time    `json:"expires_at,omitempty"`
	OccurredAt  time.Time     `json:"occurred_at"`
}
