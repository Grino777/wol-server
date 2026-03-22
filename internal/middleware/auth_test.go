package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Grino777/wol-server/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestAuthMiddleware_ValidCredentials(t *testing.T) {
	restore := setAuthConfig(t, "admin", "secret")
	defer restore()

	mw := AuthMiddleware()
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/secure", nil)
	req.SetBasicAuth("admin", "secret")
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d, want %d", rec.Code, http.StatusNoContent)
	}
	if !nextCalled {
		t.Fatal("next handler was not called")
	}
	if got := rec.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate must be empty on successful auth, got %q", got)
	}
}

func TestAuthMiddleware_Unauthorized(t *testing.T) {
	tests := []struct {
		name      string
		setupReq  func(*http.Request)
		wantCode  int
		wantWWW   string
		configU   string
		configPwd string
	}{
		{
			name:      "missing authorization header",
			setupReq:  func(_ *http.Request) {},
			wantCode:  http.StatusUnauthorized,
			wantWWW:   wwwAuthenticateHeaderValue,
			configU:   "admin",
			configPwd: "secret",
		},
		{
			name: "wrong auth scheme",
			setupReq: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer token")
			},
			wantCode:  http.StatusUnauthorized,
			wantWWW:   wwwAuthenticateHeaderValue,
			configU:   "admin",
			configPwd: "secret",
		},
		{
			name: "malformed basic auth header",
			setupReq: func(r *http.Request) {
				r.Header.Set("Authorization", "Basic !!!")
			},
			wantCode:  http.StatusUnauthorized,
			wantWWW:   wwwAuthenticateHeaderValue,
			configU:   "admin",
			configPwd: "secret",
		},
		{
			name: "invalid credentials",
			setupReq: func(r *http.Request) {
				r.SetBasicAuth("admin", "wrong")
			},
			wantCode:  http.StatusUnauthorized,
			wantWWW:   wwwAuthenticateHeaderValue,
			configU:   "admin",
			configPwd: "secret",
		},
		{
			name: "empty configured credentials are rejected",
			setupReq: func(r *http.Request) {
				r.SetBasicAuth("", "")
			},
			wantCode:  http.StatusUnauthorized,
			wantWWW:   wwwAuthenticateHeaderValue,
			configU:   "",
			configPwd: "",
		},
		{
			name: "empty configured password is rejected",
			setupReq: func(r *http.Request) {
				r.SetBasicAuth("admin", "")
			},
			wantCode:  http.StatusUnauthorized,
			wantWWW:   wwwAuthenticateHeaderValue,
			configU:   "admin",
			configPwd: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := setAuthConfig(t, tt.configU, tt.configPwd)
			defer restore()

			mw := AuthMiddleware()
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/secure", nil)
			tt.setupReq(req)
			rec := httptest.NewRecorder()

			mw(next).ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("unexpected status: got %d, want %d", rec.Code, tt.wantCode)
			}
			if got := rec.Header().Get("WWW-Authenticate"); got != tt.wantWWW {
				t.Fatalf("unexpected WWW-Authenticate header: got %q, want %q", got, tt.wantWWW)
			}
		})
	}
}

func TestAuthMiddleware_UnauthorizedLogsRequest(t *testing.T) {
	restore := setAuthConfig(t, "admin", "secret")
	defer restore()

	core, observed := observer.New(zap.WarnLevel)
	logger := zap.New(core).Sugar()

	mw := AuthMiddleware()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/secure", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.SetBasicAuth("admin", "wrong")
	req = SetRequestLogger(req, logger)
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("unexpected log entries count: got %d, want 1", len(entries))
	}

	fields := entries[0].ContextMap()
	if got := fields["method"]; got != http.MethodPost {
		t.Fatalf("unexpected method field: got %v, want %q", got, http.MethodPost)
	}
	if got := fields["path"]; got != "/secure" {
		t.Fatalf("unexpected path field: got %v, want %q", got, "/secure")
	}
	if got := fields["remote_addr"]; got != req.RemoteAddr {
		t.Fatalf("unexpected remote_addr field: got %v, want %q", got, req.RemoteAddr)
	}
	if _, exists := fields["authorization"]; exists {
		t.Fatal("authorization field must not be logged")
	}
}

func setAuthConfig(t *testing.T, username, password string) func() {
	t.Helper()

	prevUsername := configs.Username
	prevPassword := configs.UserPassword

	configs.Username = username
	configs.UserPassword = password

	return func() {
		configs.Username = prevUsername
		configs.UserPassword = prevPassword
	}
}
