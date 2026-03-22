package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Grino777/wol-server/configs"
	"github.com/Grino777/wol-server/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestSetupRoutes_LoggerMiddlewareRunsBeforeAuth(t *testing.T) {
	restore := setAuthConfig(t, "admin", "secret")
	defer restore()

	core, observed := observer.New(zap.WarnLevel)
	ctx := logging.WithLogger(context.Background(), zap.New(core).Sugar())

	a := &API{}
	router := a.SetupRoutes(ctx)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: got %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if observed.Len() == 0 {
		t.Fatal("expected unauthorized request to be logged with context logger")
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
