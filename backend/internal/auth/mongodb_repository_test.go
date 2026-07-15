package auth

import "testing"

func TestAdminEmailAllowlistUsesExactNormalizedMatch(t *testing.T) {
	repository := NewMongoUserRepository(nil, []string{"Admin@Example.com"})

	if !repository.isAdminEmail(" admin@example.com ") {
		t.Fatal("expected normalized exact email to be allowed")
	}
	for _, email := range []string{
		"notadmin@example.com",
		"admin@example.com.attacker.test",
		"admin@other.example.com",
	} {
		if repository.isAdminEmail(email) {
			t.Fatalf("email %q must not receive the admin role", email)
		}
	}
}
