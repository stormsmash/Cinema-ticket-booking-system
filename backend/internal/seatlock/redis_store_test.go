package seatlock

import "testing"

func TestLockOwnerUnderstandsBookingClaim(t *testing.T) {
	if owner := lockOwner("user-1"); owner != "user-1" {
		t.Fatalf("expected plain owner, got %q", owner)
	}
	if owner := lockOwner("booking_claim:user-1:claim-token"); owner != "user-1" {
		t.Fatalf("expected claimed owner, got %q", owner)
	}
}
