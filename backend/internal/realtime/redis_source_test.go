package realtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestParseSeatLockKey(t *testing.T) {
	screeningID := "66a000000000000000000001"

	tests := []struct {
		name string
		key  string
		ok   bool
	}{
		{name: "seat lock", key: "seat_lock:" + screeningID + ":D8", ok: true},
		{name: "session secret", key: "session:secret-token", ok: false},
		{name: "invalid screening", key: "seat_lock:not-an-id:D8", ok: false},
		{name: "missing seat", key: "seat_lock:" + screeningID + ":", ok: false},
		{name: "extra separator", key: "seat_lock:" + screeningID + ":D:8", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualScreeningID, seatID, ok := parseSeatLockKey(test.key)
			if ok != test.ok {
				t.Fatalf("expected ok=%t, got %t", test.ok, ok)
			}
			if test.ok && (actualScreeningID != screeningID || seatID != "D8") {
				t.Fatalf("unexpected parsed key: screening=%q seat=%q", actualScreeningID, seatID)
			}
		})
	}
}

func TestRedisSourceAcceptsBookedChannelEvent(t *testing.T) {
	source := &RedisSeatEventSource{}
	want := SeatEvent{
		Version:     EventVersion,
		Type:        SeatBooked,
		BookingID:   bson.NewObjectID().Hex(),
		ScreeningID: bson.NewObjectID().Hex(),
		SeatID:      "D8",
		Status:      "BOOKED",
		OccurredAt:  time.Now().UTC(),
	}
	payload, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	got, ok, err := source.toSeatEvent(context.Background(), seatEventChannel, string(payload))
	if err != nil {
		t.Fatalf("parse event: %v", err)
	}
	if !ok || got.Type != SeatBooked || got.SeatID != want.SeatID ||
		got.BookingID != want.BookingID {
		t.Fatalf("unexpected event: %#v", got)
	}
}

func TestRedisSourceRejectsInvalidBookedEvents(t *testing.T) {
	source := &RedisSeatEventSource{}
	now := time.Now().UTC()
	expiresAt := now.Add(time.Minute)
	valid := SeatEvent{
		Version:     EventVersion,
		Type:        SeatBooked,
		BookingID:   bson.NewObjectID().Hex(),
		ScreeningID: bson.NewObjectID().Hex(),
		SeatID:      "D8",
		Status:      "BOOKED",
		OccurredAt:  now,
	}

	tests := []struct {
		name   string
		mutate func(*SeatEvent)
	}{
		{name: "missing booking ID", mutate: func(event *SeatEvent) { event.BookingID = "" }},
		{name: "invalid booking ID", mutate: func(event *SeatEvent) { event.BookingID = "bad" }},
		{name: "invalid screening ID", mutate: func(event *SeatEvent) { event.ScreeningID = "bad" }},
		{name: "blank seat", mutate: func(event *SeatEvent) { event.SeatID = " " }},
		{name: "long seat", mutate: func(event *SeatEvent) { event.SeatID = "THIS-SEAT-ID-IS-TOO-LONG" }},
		{name: "wrong version", mutate: func(event *SeatEvent) { event.Version = 2 }},
		{name: "wrong type", mutate: func(event *SeatEvent) { event.Type = SeatLocked }},
		{name: "wrong status", mutate: func(event *SeatEvent) { event.Status = "AVAILABLE" }},
		{name: "missing time", mutate: func(event *SeatEvent) { event.OccurredAt = time.Time{} }},
		{name: "booking expiry", mutate: func(event *SeatEvent) { event.ExpiresAt = &expiresAt }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := valid
			test.mutate(&event)
			payload, err := json.Marshal(event)
			if err != nil {
				t.Fatalf("marshal event: %v", err)
			}

			_, ok, err := source.toSeatEvent(context.Background(), seatEventChannel, string(payload))
			if err != nil {
				t.Fatalf("parse event: %v", err)
			}
			if ok {
				t.Fatal("expected invalid event to be ignored")
			}
		})
	}
}
