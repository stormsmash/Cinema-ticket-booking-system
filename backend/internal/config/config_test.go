package config

import "testing"

func TestLoadUsesDefaults(t *testing.T) {
	for _, key := range []string{
		"PORT",
		"GIN_MODE",
		"MONGO_URI",
		"MONGO_DATABASE",
		"REDIS_ADDRESS",
		"REDIS_PASSWORD",
		"REDIS_DB",
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
}

func TestLoadRejectsInvalidRedisDB(t *testing.T) {
	t.Setenv("REDIS_DB", "not-a-number")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid REDIS_DB to return an error")
	}
}
