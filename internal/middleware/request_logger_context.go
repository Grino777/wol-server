package middleware

import (
	"net/http"

	"github.com/Grino777/wol-server/pkg/logging"
	"go.uber.org/zap"
)

// ContextLogger
func ContextLogger(logger *zap.SugaredLogger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, SetRequestLogger(r, logger))
		})
	}
}

// SetRequestLogger returns a cloned request with logger in context.
func SetRequestLogger(r *http.Request, logger *zap.SugaredLogger) *http.Request {
	if r == nil {
		return nil
	}

	ctx := r.Context()
	if logger == nil {
		logger = logging.FromContext(ctx)
	}

	return r.WithContext(logging.WithLogger(ctx, logger))
}

// GetRequestLogger returns logger from request context with safe fallback.
func GetRequestLogger(r *http.Request) *zap.SugaredLogger {
	if r == nil {
		return logging.DefaultLogger()
	}
	return logging.FromContext(r.Context())
}
