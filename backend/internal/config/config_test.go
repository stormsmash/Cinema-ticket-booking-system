package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	for _, key := range []string{
		"PORT",
		"GIN_MODE",
		"MONGO_URI",
		"MONGO_DATABASE",
		"REDIS_ADDRESS",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"GOOGLE_CLIENT_ID",
		"GOOGLE_CLIENT_SECRET",
		"GOOGLE_REDIRECT_URL",
		"FRONTEND_URL",
		"SESSION_TTL",
		"SEAT_LOCK_TTL",
		"COOKIE_SECURE",
	} {
		t.Setenv(key, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %q, got %q", defaultPort, cfg.Port)
	}
	if cfg.MongoURI != defaultMongoURI {
		t.Fatalf("expected default MongoDB URI %q, got %q", defaultMongoURI, cfg.MongoURI)
	}
	if cfg.RedisAddress != defaultRedisAddress {
		t.Fatalf("expected default Redis address %q, got %q", defaultRedisAddress, cfg.RedisAddress)
	}
	if cfg.RedisDB != 0 {
		t.Fatalf("expected default Redis DB 0, got %d", cfg.RedisDB)
	}
	if cfg.SessionTTL != 24*time.Hour {
		t.Fatalf("expected default session TTL 24h, got %q", cfg.SessionTTL)
	}
	if cfg.SeatLockTTL != 10*time.Minute {
		t.Fatalf("expected default seat lock TTL 10m, got %q", cfg.SeatLockTTL)
	}
	if cfg.GoogleOAuthEnabled() {
		t.Fatal("expected Google OAuth to be disabled without credentials")
	}
}

func TestLoadRejectsInvalidRedisDB(t *testing.T) {
	t.Setenv("REDIS_DB", "not-a-number")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid REDIS_DB to return an error")
	}
}

func TestLoadRejectsInvalidSessionTTL(t *testing.T) {
	t.Setenv("SESSION_TTL", "tomorrow")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid SESSION_TTL to return an error")
	}
}

func TestLoadRejectsInvalidSeatLockTTL(t *testing.T) {
	t.Setenv("SEAT_LOCK_TTL", "0s")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid SEAT_LOCK_TTL to return an error")
	}
}

func TestGoogleOAuthEnabledRequiresBothCredentials(t *testing.T) {
	config := Config{GoogleClientID: "client", GoogleClientSecret: "secret"}

	if !config.GoogleOAuthEnabled() {
		t.Fatal("expected Google OAuth to be enabled with both credentials")
	}
}
