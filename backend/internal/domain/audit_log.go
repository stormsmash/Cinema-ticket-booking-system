package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuditEvent string

const (
	AuditEventBookingSuccess AuditEvent = "BOOKING_SUCCESS"
)

type AuditLog struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Event       AuditEvent    `bson:"event"`
	BookingID   bson.ObjectID `bson:"booking_id"`
	UserID      string        `bson:"user_id"`
	ScreeningID bson.ObjectID `bson:"screening_id"`
	SeatID      string        `bson:"seat_id"`
	CreatedAt   time.Time     `bson:"created_at"`
}
