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
	if !ok || got.Type != SeatBooked || got.SeatID != want.SeatID {
		t.Fatalf("unexpected event: %#v", got)
	}
}
