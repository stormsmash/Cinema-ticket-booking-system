package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuditEvent string

const (
	AuditEventBookingSuccess AuditEvent = "BOOKING_SUCCESS"
	AuditEventBookingTimeout AuditEvent = "BOOKING_TIMEOUT"
	AuditEventSeatReleased   AuditEvent = "SEAT_RELEASED"
	AuditEventSystemError    AuditEvent = "SYSTEM_ERROR"
)

type AuditLog struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Event       AuditEvent    `bson:"event"`
	BookingID   bson.ObjectID `bson:"booking_id,omitempty"`
	UserID      string        `bson:"user_id,omitempty"`
	ScreeningID bson.ObjectID `bson:"screening_id,omitempty"`
	SeatID      string        `bson:"seat_id,omitempty"`
	Message     string        `bson:"message,omitempty"`
	CreatedAt   time.Time     `bson:"created_at"`
}
