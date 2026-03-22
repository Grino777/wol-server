package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/Grino777/wol-server/configs"
)

const unauthorizedMessage = "unauthorized"
const wwwAuthenticateHeaderValue = `Basic realm="restricted", charset="UTF-8"`

func AuthMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok || !isAuthorized(username, password) {
				GetRequestLogger(r).Warnw("unauthorized request",
					"method", r.Method,
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				w.Header().Set("WWW-Authenticate", wwwAuthenticateHeaderValue)
				writeJSONError(w, http.StatusUnauthorized, unauthorizedMessage)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMidlleware is kept for backward compatibility with historical typo.
func AuthMidlleware() Middleware {
	return AuthMiddleware()
}

func isAuthorized(username, password string) bool {
	if configs.Username == "" || configs.UserPassword == "" {
		return false
	}

	return secureCompare(username, configs.Username) && secureCompare(password, configs.UserPassword)
}

func secureCompare(input, expected string) bool {
	inputHash := sha256.Sum256([]byte(input))
	expectedHash := sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(inputHash[:], expectedHash[:]) == 1
}
