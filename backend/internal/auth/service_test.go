package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

type providerStub struct {
	profile GoogleProfile
}

func (stub providerStub) AuthorizationURL(state string) string {
	return "https://accounts.example.test?state=" + state
}

func (stub providerStub) FetchProfile(context.Context, string) (GoogleProfile, error) {
	return stub.profile, nil
}

type userRepositoryStub struct {
	user domain.User
}

func (stub userRepositoryStub) UpsertGoogleUser(
	context.Context,
	GoogleProfile,
) (domain.User, error) {
	return stub.user, nil
}

func (stub userRepositoryStub) FindByID(context.Context, string) (domain.User, error) {
	return stub.user, nil
}

type sessionStoreStub struct {
	token  string
	userID string
}

func (stub sessionStoreStub) Create(context.Context, string, time.Duration) (string, error) {
	return stub.token, nil
}

func (stub sessionStoreStub) UserID(context.Context, string) (string, error) {
	if stub.userID == "" {
		return "", ErrSessionNotFound
	}
	return stub.userID, nil
}

func (stub sessionStoreStub) Delete(context.Context, string) error {
	return nil
}

func TestBeginLoginReturnsNotConfigured(t *testing.T) {
	service := NewService(providerStub{}, userRepositoryStub{}, sessionStoreStub{}, time.Hour, false)

	if _, _, err := service.BeginLogin(); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestCompleteLoginCreatesSessionForVerifiedGoogleUser(t *testing.T) {
	user := domain.User{ID: bson.NewObjectID(), Email: "viewer@example.com", Name: "Viewer"}
	service := NewService(
		providerStub{profile: GoogleProfile{
			Subject:       "google-subject",
			Email:         user.Email,
			EmailVerified: true,
			Name:          user.Name,
		}},
		userRepositoryStub{user: user},
		sessionStoreStub{token: "session-token"},
		24*time.Hour,
		true,
	)

	token, result, err := service.CompleteLogin(context.Background(), "authorization-code")
	if err != nil {
		t.Fatalf("complete login: %v", err)
	}
	if token != "session-token" {
		t.Fatalf("expected session token, got %q", token)
	}
	if result.ID != user.ID {
		t.Fatalf("expected user %s, got %s", user.ID.Hex(), result.ID.Hex())
	}
}

func TestCompleteLoginRejectsUnverifiedEmail(t *testing.T) {
	service := NewService(
		providerStub{profile: GoogleProfile{Subject: "google-subject", Email: "user@example.com"}},
		userRepositoryStub{},
		sessionStoreStub{},
		time.Hour,
		true,
	)

	if _, _, err := service.CompleteLogin(context.Background(), "code"); err == nil {
		t.Fatal("expected unverified profile to be rejected")
	}
}

func TestCurrentUserRejectsMissingSession(t *testing.T) {
	service := NewService(
		providerStub{},
		userRepositoryStub{},
		sessionStoreStub{},
		time.Hour,
		true,
	)

	if _, err := service.CurrentUser(context.Background(), "missing"); !errors.Is(err, ErrNotAuthenticated) {
		t.Fatalf("expected ErrNotAuthenticated, got %v", err)
	}
}
