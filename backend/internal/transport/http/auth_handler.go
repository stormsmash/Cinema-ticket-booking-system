package httptransport

import (
	"context"
	"crypto/subtle"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/auth"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
)

const (
	sessionCookieName  = "cinema_session"
	stateCookieName    = "cinema_oauth_state"
	authUserContextKey = "auth_user"
	stateTTL           = 10 * time.Minute
)

type AuthService interface {
	Enabled() bool
	BeginLogin() (string, string, error)
	CompleteLogin(context.Context, string) (string, domain.User, error)
	CurrentUser(context.Context, string) (domain.User, error)
	Logout(context.Context, string) error
}

type AuthHandlerConfig struct {
	FrontendURL  string
	SessionTTL   time.Duration
	CookieSecure bool
}

type authHandler struct {
	service AuthService
	config  AuthHandlerConfig
}

type authConfigResponse struct {
	GoogleEnabled bool `json:"google_enabled"`
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Role      string `json:"role"`
}

func (handler *authHandler) requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := optionalUser(c)
		if !exists || user.Role != domain.UserRoleAdmin {
			writeError(c, http.StatusForbidden, "ADMIN_REQUIRED", "Administrator access is required")
			c.Abort()
			return
		}

		c.Next()
	}
}

func newAuthHandler(service AuthService, config AuthHandlerConfig) *authHandler {
	return &authHandler{service: service, config: config}
}

func (handler *authHandler) configuration(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": authConfigResponse{GoogleEnabled: handler.service.Enabled()},
	})
}

func (handler *authHandler) google(c *gin.Context) {
	state, authorizationURL, err := handler.service.BeginLogin()
	if errors.Is(err, auth.ErrNotConfigured) {
		writeError(
			c,
			http.StatusServiceUnavailable,
			"AUTH_NOT_CONFIGURED",
			"Google sign-in is not configured",
		)
		return
	}
	if err != nil {
		log.Printf("begin Google login: %v", err)
		writeError(c, http.StatusInternalServerError, "AUTH_START_FAILED", "Unable to start sign-in")
		return
	}

	handler.setCookie(c, stateCookieName, state, stateTTL, "/api/v1/auth/google/callback")
	c.Redirect(http.StatusFound, authorizationURL)
}

func (handler *authHandler) googleCallback(c *gin.Context) {
	handler.clearCookie(c, stateCookieName, "/api/v1/auth/google/callback")

	if c.Query("error") != "" {
		handler.redirectWithError(c, "access_denied")
		return
	}

	stateCookie, err := c.Cookie(stateCookieName)
	if err != nil || !sameValue(stateCookie, c.Query("state")) {
		handler.redirectWithError(c, "invalid_state")
		return
	}

	code := c.Query("code")
	if code == "" {
		handler.redirectWithError(c, "missing_code")
		return
	}

	sessionToken, _, err := handler.service.CompleteLogin(c.Request.Context(), code)
	if err != nil {
		log.Printf("complete Google login: %v", err)
		handler.redirectWithError(c, "login_failed")
		return
	}

	handler.setCookie(c, sessionCookieName, sessionToken, handler.config.SessionTTL, "/")
	c.Redirect(http.StatusSeeOther, handler.config.FrontendURL)
}

func (handler *authHandler) me(c *gin.Context) {
	user := c.MustGet(authUserContextKey).(domain.User)
	c.JSON(http.StatusOK, gin.H{"data": toUserResponse(user)})
}

func (handler *authHandler) requireAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken, err := c.Cookie(sessionCookieName)
		if err != nil {
			writeError(c, http.StatusUnauthorized, "NOT_AUTHENTICATED", "Sign in is required")
			c.Abort()
			return
		}

		user, err := handler.service.CurrentUser(c.Request.Context(), sessionToken)
		if errors.Is(err, auth.ErrNotAuthenticated) {
			handler.clearCookie(c, sessionCookieName, "/")
			writeError(c, http.StatusUnauthorized, "NOT_AUTHENTICATED", "Sign in is required")
			c.Abort()
			return
		}
		if err != nil {
			log.Printf("load current user: %v", err)
			writeError(c, http.StatusInternalServerError, "AUTH_CHECK_FAILED", "Unable to check session")
			c.Abort()
			return
		}

		c.Set(authUserContextKey, user)
		c.Next()
	}
}

func (handler *authHandler) optionalAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionToken, err := c.Cookie(sessionCookieName)
		if err != nil {
			c.Next()
			return
		}

		user, err := handler.service.CurrentUser(c.Request.Context(), sessionToken)
		if errors.Is(err, auth.ErrNotAuthenticated) {
			handler.clearCookie(c, sessionCookieName, "/")
			c.Next()
			return
		}
		if err != nil {
			log.Printf("load optional current user: %v", err)
			writeError(c, http.StatusInternalServerError, "AUTH_CHECK_FAILED", "Unable to check session")
			c.Abort()
			return
		}

		c.Set(authUserContextKey, user)
		c.Next()
	}
}

func authenticatedUser(c *gin.Context) domain.User {
	return c.MustGet(authUserContextKey).(domain.User)
}

func optionalUser(c *gin.Context) (domain.User, bool) {
	value, exists := c.Get(authUserContextKey)
	if !exists {
		return domain.User{}, false
	}

	user, ok := value.(domain.User)
	return user, ok
}

func (handler *authHandler) logout(c *gin.Context) {
	sessionToken, _ := c.Cookie(sessionCookieName)
	if err := handler.service.Logout(c.Request.Context(), sessionToken); err != nil {
		log.Printf("logout: %v", err)
		writeError(c, http.StatusInternalServerError, "LOGOUT_FAILED", "Unable to sign out")
		return
	}

	handler.clearCookie(c, sessionCookieName, "/")
	c.Status(http.StatusNoContent)
}

func (handler *authHandler) setCookie(
	c *gin.Context,
	name string,
	value string,
	ttl time.Duration,
	path string,
) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Expires:  time.Now().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: true,
		Secure:   handler.config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (handler *authHandler) clearCookie(c *gin.Context, name, path string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   handler.config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (handler *authHandler) redirectWithError(c *gin.Context, code string) {
	redirectURL, err := url.Parse(handler.config.FrontendURL)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	query := redirectURL.Query()
	query.Set("auth_error", code)
	redirectURL.RawQuery = query.Encode()
	c.Redirect(http.StatusSeeOther, redirectURL.String())
}

func sameValue(left, right string) bool {
	if left == "" || len(left) != len(right) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func toUserResponse(user domain.User) userResponse {
	return userResponse{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Role:      string(user.Role),
	}
}
