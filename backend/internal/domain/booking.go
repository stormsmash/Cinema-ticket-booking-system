package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BookingStatus string

const (
	BookingStatusHolding   BookingStatus = "HOLDING"
	BookingStatusBooked    BookingStatus = "BOOKED"
	BookingStatusTimedOut  BookingStatus = "TIMED_OUT"
	BookingStatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	UserID      string        `bson:"user_id"`
	ScreeningID bson.ObjectID `bson:"screening_id"`
	SeatID      string        `bson:"seat_id"`
	PriceBaht   int           `bson:"price_baht"`
	Status      BookingStatus `bson:"status"`
	ExpiresAt   *time.Time    `bson:"expires_at,omitempty"`
	CreatedAt   time.Time     `bson:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"`
}
