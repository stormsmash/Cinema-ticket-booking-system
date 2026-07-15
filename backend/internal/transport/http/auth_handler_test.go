package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/health"
)

type authServiceStub struct {
	enabled      bool
	beginState   string
	beginURL     string
	sessionToken string
	user         domain.User
	completed    bool
}

func (stub *authServiceStub) Enabled() bool {
	return stub.enabled
}

func (stub *authServiceStub) BeginLogin() (string, string, error) {
	return stub.beginState, stub.beginURL, nil
}

func (stub *authServiceStub) CompleteLogin(
	context.Context,
	string,
) (string, domain.User, error) {
	stub.completed = true
	return stub.sessionToken, stub.user, nil
}

func (stub *authServiceStub) CurrentUser(context.Context, string) (domain.User, error) {
	return stub.user, nil
}

func (stub *authServiceStub) Logout(context.Context, string) error {
	return nil
}

func authTestRouter(service AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(Dependencies{
		Readiness: readinessFunc(func(context.Context) health.Report {
			return health.Report{Status: health.StatusReady}
		}),
		Screenings: screeningServiceStub{},
		Auth:       service,
		AuthConfig: AuthHandlerConfig{
			FrontendURL: "http://localhost:3000",
			SessionTTL:  24 * time.Hour,
		},
	})
}

func TestAuthConfigReportsGoogleStatus(t *testing.T) {
	router := authTestRouter(&authServiceStub{enabled: true})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/config", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"google_enabled":true`) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestGoogleLoginSetsStateCookieAndRedirects(t *testing.T) {
	service := &authServiceStub{
		enabled:    true,
		beginState: "oauth-state",
		beginURL:   "https://accounts.example.test/login",
	}
	router := authTestRouter(service)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, recorder.Code)
	}
	if recorder.Header().Get("Location") != service.beginURL {
		t.Fatalf("unexpected redirect: %s", recorder.Header().Get("Location"))
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != stateCookieName || !cookies[0].HttpOnly {
		t.Fatalf("unexpected state cookie: %#v", cookies)
	}
}

func TestGoogleCallbackRejectsMismatchedState(t *testing.T) {
	service := &authServiceStub{}
	router := authTestRouter(service)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/auth/google/callback?state=wrong&code=code",
		nil,
	)
	request.AddCookie(&http.Cookie{Name: stateCookieName, Value: "expected"})

	router.ServeHTTP(recorder, request)

	if service.completed {
		t.Fatal("OAuth code must not be exchanged when state does not match")
	}
	if recorder.Code != http.StatusSeeOther ||
		!strings.Contains(recorder.Header().Get("Location"), "auth_error=invalid_state") {
		t.Fatalf("unexpected redirect: %s", recorder.Header().Get("Location"))
	}
}

func TestGoogleCallbackCreatesSessionCookie(t *testing.T) {
	service := &authServiceStub{
		sessionToken: "session-token",
		user: domain.User{
			ID:    bson.NewObjectID(),
			Email: "viewer@example.com",
			Name:  "Viewer",
		},
	}
	router := authTestRouter(service)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/auth/google/callback?state=expected&code=code",
		nil,
	)
	request.AddCookie(&http.Cookie{Name: stateCookieName, Value: "expected"})

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusSeeOther || !service.completed {
		t.Fatalf("expected successful callback, status=%d", recorder.Code)
	}
	var sessionCookie *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == sessionCookieName {
			sessionCookie = cookie
		}
	}
	if sessionCookie == nil || sessionCookie.Value != service.sessionToken || !sessionCookie.HttpOnly {
		t.Fatalf("unexpected session cookie: %#v", sessionCookie)
	}
}

func TestMeRequiresSessionCookie(t *testing.T) {
	router := authTestRouter(&authServiceStub{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	service := &authServiceStub{
		user: domain.User{
			ID:    bson.NewObjectID(),
			Email: "viewer@example.com",
			Name:  "Cinema Viewer",
		},
	}
	router := authTestRouter(service)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), service.user.Email) {
		t.Fatalf("unexpected response: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
