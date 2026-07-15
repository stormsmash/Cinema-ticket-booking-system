package realtime

import (
	"errors"
	"testing"
	"time"
)

func TestHubPublishesOnlyToMatchingScreening(t *testing.T) {
	hub := NewHub(10)
	first, unsubscribeFirst, err := hub.Subscribe("screening-a")
	if err != nil {
		t.Fatalf("subscribe first client: %v", err)
	}
	defer unsubscribeFirst()
	second, unsubscribeSecond, err := hub.Subscribe("screening-b")
	if err != nil {
		t.Fatalf("subscribe second client: %v", err)
	}
	defer unsubscribeSecond()

	event := SeatEvent{ScreeningID: "screening-a", SeatID: "A1"}
	hub.Publish(event)

	select {
	case received := <-first:
		if received.SeatID != event.SeatID {
			t.Fatalf("expected seat %q, got %q", event.SeatID, received.SeatID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("matching subscriber did not receive the event")
	}

	select {
	case <-second:
		t.Fatal("different screening must not receive the event")
	case <-time.After(20 * time.Millisecond):
	}
}

func TestHubEnforcesSubscriberLimitAndCloses(t *testing.T) {
	hub := NewHub(1)
	events, unsubscribe, err := hub.Subscribe("screening-a")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer unsubscribe()

	if _, _, err := hub.Subscribe("screening-a"); !errors.Is(err, ErrHubFull) {
		t.Fatalf("expected ErrHubFull, got %v", err)
	}

	hub.Close()
	if _, open := <-events; open {
		t.Fatal("expected subscriber channel to close with the hub")
	}
	if _, _, err := hub.Subscribe("screening-b"); !errors.Is(err, ErrHubClosed) {
		t.Fatalf("expected ErrHubClosed, got %v", err)
	}
}
