package mongodb

import (
	"testing"
	"time"
)

func TestScreeningSeedsHaveStableLayouts(t *testing.T) {
	seeds := screeningSeeds(time.Date(2026, time.July, 14, 12, 0, 0, 0, time.UTC))

	if len(seeds) != 2 {
		t.Fatalf("expected 2 screenings, got %d", len(seeds))
	}
	if len(seeds[0].Seats) != 50 {
		t.Fatalf("expected first screening to have 50 seats, got %d", len(seeds[0].Seats))
	}
	if seeds[0].Seats[0].ID != "A1" || seeds[0].Seats[49].ID != "E10" {
		t.Fatalf("unexpected seat layout: first=%q last=%q", seeds[0].Seats[0].ID, seeds[0].Seats[49].ID)
	}
}
