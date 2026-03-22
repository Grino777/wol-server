package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestContextLogger_PutsLoggerIntoRequestContext(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	expectedLogger := zap.New(core).Sugar()

	var gotLogger *zap.SugaredLogger
	mw := ContextLogger(expectedLogger)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLogger = GetRequestLogger(r)
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mw(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d, want %d", rec.Code, http.StatusNoContent)
	}
	if gotLogger != expectedLogger {
		t.Fatal("logger from request context does not match expected logger")
	}
}
