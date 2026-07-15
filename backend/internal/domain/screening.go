package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "AVAILABLE"
	SeatStatusLocked    SeatStatus = "LOCKED"
	SeatStatusBooked    SeatStatus = "BOOKED"
)

type Movie struct {
	Title           string `bson:"title"`
	DurationMinutes int    `bson:"duration_minutes"`
}

type Auditorium struct {
	Name        string `bson:"name"`
	Rows        int    `bson:"rows"`
	SeatsPerRow int    `bson:"seats_per_row"`
}

type Seat struct {
	ID     string     `bson:"id"`
	Row    string     `bson:"row"`
	Number int        `bson:"number"`
	Status SeatStatus `bson:"status"`
}

type Screening struct {
	ID         bson.ObjectID `bson:"_id,omitempty"`
	Movie      Movie         `bson:"movie"`
	Auditorium Auditorium    `bson:"auditorium"`
	StartsAt   time.Time     `bson:"starts_at"`
	Seats      []Seat        `bson:"seats"`
	CreatedAt  time.Time     `bson:"created_at"`
	UpdatedAt  time.Time     `bson:"updated_at"`
}
