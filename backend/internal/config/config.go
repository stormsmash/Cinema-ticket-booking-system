package config

import (
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort              = "8080"
	defaultGinMode           = "debug"
	defaultMongoURI          = "mongodb://localhost:27017/cinema?replicaSet=rs0&directConnection=true"
	defaultMongoDatabase     = "cinema"
	defaultRedisAddress      = "localhost:6379"
	defaultRedisDB           = "0"
	defaultGoogleRedirectURL = "http://localhost:3000/api/v1/auth/google/callback"
	defaultFrontendURL       = "http://localhost:3000"
	defaultSessionTTL        = "24h"
	defaultSeatLockTTL       = "5m"
	defaultCookieSecure      = "false"
)

type Config struct {
	Port               string
	GinMode            string
	MongoURI           string
	MongoDatabase      string
	RedisAddress       string
	RedisPassword      string
	RedisDB            int
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendURL        string
	SessionTTL         time.Duration
	SeatLockTTL        time.Duration
	CookieSecure       bool
	AdminEmails        []string
}

func Load() (Config, error) {
	redisDB, err := strconv.Atoi(valueOrDefault("REDIS_DB", defaultRedisDB))
	if err != nil || redisDB < 0 {
		return Config{}, fmt.Errorf("REDIS_DB must be a non-negative integer")
	}

	sessionTTL, err := time.ParseDuration(valueOrDefault("SESSION_TTL", defaultSessionTTL))
	if err != nil || sessionTTL <= 0 {
		return Config{}, fmt.Errorf("SESSION_TTL must be a positive duration")
	}

	seatLockTTL, err := time.ParseDuration(valueOrDefault("SEAT_LOCK_TTL", defaultSeatLockTTL))
	if err != nil || seatLockTTL <= 0 {
		return Config{}, fmt.Errorf("SEAT_LOCK_TTL must be a positive duration")
	}

	cookieSecure, err := strconv.ParseBool(valueOrDefault("COOKIE_SECURE", defaultCookieSecure))
	if err != nil {
		return Config{}, fmt.Errorf("COOKIE_SECURE must be true or false")
	}

	adminEmails, err := parseAdminEmails(os.Getenv("ADMIN_EMAILS"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:               valueOrDefault("PORT", defaultPort),
		GinMode:            valueOrDefault("GIN_MODE", defaultGinMode),
		MongoURI:           valueOrDefault("MONGO_URI", defaultMongoURI),
		MongoDatabase:      valueOrDefault("MONGO_DATABASE", defaultMongoDatabase),
		RedisAddress:       valueOrDefault("REDIS_ADDRESS", defaultRedisAddress),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		RedisDB:            redisDB,
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  valueOrDefault("GOOGLE_REDIRECT_URL", defaultGoogleRedirectURL),
		FrontendURL:        valueOrDefault("FRONTEND_URL", defaultFrontendURL),
		SessionTTL:         sessionTTL,
		SeatLockTTL:        seatLockTTL,
		CookieSecure:       cookieSecure,
		AdminEmails:        adminEmails,
	}, nil
}

func parseAdminEmails(value string) ([]string, error) {
	seen := make(map[string]struct{})
	emails := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		email := strings.ToLower(strings.TrimSpace(item))
		if email == "" {
			continue
		}
		parsed, err := mail.ParseAddress(email)
		if err != nil || parsed.Address != email {
			return nil, fmt.Errorf("ADMIN_EMAILS must contain comma-separated email addresses")
		}
		if _, exists := seen[email]; exists {
			continue
		}
		seen[email] = struct{}{}
		emails = append(emails, email)
	}

	return emails, nil
}

func (config Config) GoogleOAuthEnabled() bool {
	return config.GoogleClientID != "" && config.GoogleClientSecret != ""
}

func valueOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
