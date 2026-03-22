package middleware

import (
	"fmt"
	"net/http"

	"github.com/felixge/httpsnoop"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// // PopulateLogger populates the logger onto the context.
// func PopulateLogger(originalLogger *zap.SugaredLogger) Middleware {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			ctx := r.Context()

// 			logger := originalLogger

// 			// Only override the logger if it's the default logger. This is only used
// 			// for testing and is intentionally a strict object equality check because
// 			// the default logger is a global default in the logger package.
// 			if existing := logging.FromContext(ctx); existing == logging.DefaultLogger() {
// 				logger = existing
// 			}

// 			// If there's a request ID, set that on the logger.
// 			if rid := RequestIDFromContext(ctx); rid != "" {
// 				logger = logger.With("request_id", rid)
// 			}

// 			ctx = logging.WithLogger(ctx, logger)
// 			r = r.Clone(ctx)

// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

// ANSI foreground color codes used by RequestLogger to colorize terminal output.
const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiCyan   = "\x1b[36m"
)

// colorizeMethod wraps the HTTP method in an ANSI color:
// GET - green, POST - yellow, PUT - cyan, DELETE - red, others - no color.
func colorizeMethod(method string) string {
	var color string
	switch method {
	case http.MethodGet:
		color = ansiGreen
	case http.MethodPost:
		color = ansiYellow
	case http.MethodPut:
		color = ansiCyan
	case http.MethodDelete:
		color = ansiRed
	default:
		return method
	}
	return color + method + ansiReset
}

// colorizeStatus wraps the HTTP status code in an ANSI color based on its range:
// green for 2xx, yellow for 3xx-4xx, red for 5xx+.
func colorizeStatus(code int) string {
	var color string
	switch {
	case code >= 200 && code < 300:
		color = ansiGreen
	case code >= 300 && code < 500:
		color = ansiYellow
	default:
		color = ansiRed
	}
	return fmt.Sprintf("%s%d%s", color, code, ansiReset)
}

// RequestLogger returns a middleware that logs each HTTP request as structured
// JSON via the provided zap logger. It captures method, URL, protocol, client
// IP, chi request ID, response status, bytes written, and duration.
//
// Must be placed after chimw.RealIP and chimw.RequestID in the middleware chain.
func RequestLogger(logger *zap.SugaredLogger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}

			requestID := chimw.GetReqID(r.Context())

			var m httpsnoop.Metrics
			defer func() {
				url := fmt.Sprintf("%s%s://%s%s%s", ansiYellow, scheme, r.Host, r.RequestURI, ansiReset)

				logger.Debugw("http request",
					"request_id", requestID,
					"method", colorizeMethod(r.Method),
					"url", url,
					"proto", r.Proto,
					"remote_addr", r.RemoteAddr,
					"user_agent", r.UserAgent(),
					"status", colorizeStatus(m.Code),
					"bytes_written", m.Written,
					"duration", m.Duration.String(),
				)
			}()

			m = httpsnoop.CaptureMetrics(next, w, r)
		})
	}
}
