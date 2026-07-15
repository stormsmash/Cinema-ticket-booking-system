package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

type GoogleProvider struct {
	config *oauth2.Config
}

func NewGoogleProvider(clientID, clientSecret, redirectURL string) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (provider *GoogleProvider) AuthorizationURL(state string) string {
	return provider.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (provider *GoogleProvider) FetchProfile(
	ctx context.Context,
	code string,
) (GoogleProfile, error) {
	token, err := provider.config.Exchange(ctx, code)
	if err != nil {
		return GoogleProfile{}, fmt.Errorf("exchange authorization code: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return GoogleProfile{}, fmt.Errorf("create user info request: %w", err)
	}

	response, err := provider.config.Client(ctx, token).Do(request)
	if err != nil {
		return GoogleProfile{}, fmt.Errorf("request user info: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return GoogleProfile{}, fmt.Errorf("user info returned status %d", response.StatusCode)
	}

	var payload struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return GoogleProfile{}, fmt.Errorf("decode user info: %w", err)
	}

	return GoogleProfile{
		Subject:       payload.Subject,
		Email:         payload.Email,
		EmailVerified: payload.EmailVerified,
		Name:          payload.Name,
		AvatarURL:     payload.Picture,
	}, nil
}
