package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/Grino777/wol-server/configs"
)

const requestBodyTooLargeMessage = "request body too large"

type errorResponse struct {
	Error string `json:"error"`
}

// MaxBytes limits request body size using Content-Length precheck and
// http.MaxBytesReader for runtime enforcement while reading body.
func MaxBytes() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > configs.MaxWakeRequestBodyBytes {
				writeJSONError(w, http.StatusRequestEntityTooLarge, requestBodyTooLargeMessage)
				if r.Body != nil {
					_ = r.Body.Close()
				}
				return
			}

			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, configs.MaxWakeRequestBodyBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: message})
}
