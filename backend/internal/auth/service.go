package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

var (
	ErrNotConfigured    = errors.New("Google OAuth is not configured")
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrSessionNotFound  = errors.New("session not found")
	ErrUserNotFound     = errors.New("user not found")
)

type GoogleProfile struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	AvatarURL     string
}

type OAuthProvider interface {
	AuthorizationURL(state string) string
	FetchProfile(context.Context, string) (GoogleProfile, error)
}

type UserRepository interface {
	UpsertGoogleUser(context.Context, GoogleProfile) (domain.User, error)
	FindByID(context.Context, string) (domain.User, error)
}

type SessionStore interface {
	Create(context.Context, string, time.Duration) (string, error)
	UserID(context.Context, string) (string, error)
	Delete(context.Context, string) error
}

type Service struct {
	provider   OAuthProvider
	users      UserRepository
	sessions   SessionStore
	sessionTTL time.Duration
	enabled    bool
}

func NewService(
	provider OAuthProvider,
	users UserRepository,
	sessions SessionStore,
	sessionTTL time.Duration,
	enabled bool,
) *Service {
	return &Service{
		provider:   provider,
		users:      users,
		sessions:   sessions,
		sessionTTL: sessionTTL,
		enabled:    enabled,
	}
}

func (service *Service) Enabled() bool {
	return service.enabled
}

func (service *Service) BeginLogin() (string, string, error) {
	if !service.enabled {
		return "", "", ErrNotConfigured
	}

	state, err := randomToken()
	if err != nil {
		return "", "", fmt.Errorf("generate OAuth state: %w", err)
	}

	return state, service.provider.AuthorizationURL(state), nil
}

func (service *Service) CompleteLogin(
	ctx context.Context,
	code string,
) (string, domain.User, error) {
	if !service.enabled {
		return "", domain.User{}, ErrNotConfigured
	}

	profile, err := service.provider.FetchProfile(ctx, code)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("fetch Google profile: %w", err)
	}
	if profile.Subject == "" || profile.Email == "" || !profile.EmailVerified {
		return "", domain.User{}, errors.New("Google profile is incomplete or unverified")
	}

	user, err := service.users.UpsertGoogleUser(ctx, profile)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("save user: %w", err)
	}

	sessionToken, err := service.sessions.Create(ctx, user.ID.Hex(), service.sessionTTL)
	if err != nil {
		return "", domain.User{}, fmt.Errorf("create session: %w", err)
	}

	return sessionToken, user, nil
}

func (service *Service) CurrentUser(ctx context.Context, sessionToken string) (domain.User, error) {
	if sessionToken == "" {
		return domain.User{}, ErrNotAuthenticated
	}

	userID, err := service.sessions.UserID(ctx, sessionToken)
	if errors.Is(err, ErrSessionNotFound) {
		return domain.User{}, ErrNotAuthenticated
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("load session: %w", err)
	}

	user, err := service.users.FindByID(ctx, userID)
	if errors.Is(err, ErrUserNotFound) {
		return domain.User{}, ErrNotAuthenticated
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("load user: %w", err)
	}

	return user, nil
}

func (service *Service) Logout(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return nil
	}

	if err := service.sessions.Delete(ctx, sessionToken); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

func randomToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
