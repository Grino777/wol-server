// Package jsonutil provides common utilities for properly handling JSON payloads in HTTP body.
package jsonutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Grino777/wol-server/configs"
)

// UnmarshalRequest provides a common implementation of JSON unmarshalling with well defined error handling.
func UnmarshalRequest(w http.ResponseWriter, r *http.Request, data interface{}) (int, error) {
	if t := r.Header.Get("content-type"); len(t) < 16 || t[:16] != "application/json" {
		return http.StatusUnsupportedMediaType, fmt.Errorf("content-type is not application/json")
	}

	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, configs.MaxWakeRequestBodyBytes)

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	if err := d.Decode(&data); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			return http.StatusBadRequest, fmt.Errorf("malformed json at position %v", syntaxErr.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return http.StatusBadRequest, fmt.Errorf("malformed json")
		case errors.As(err, &unmarshalError):
			return http.StatusBadRequest, fmt.Errorf("invalid value %v at position %v", unmarshalError.Field, unmarshalError.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return http.StatusBadRequest, fmt.Errorf("unknown field %s", fieldName)
		case errors.Is(err, io.EOF):
			return http.StatusBadRequest, fmt.Errorf("body must not be empty")
		case err.Error() == "http: request body too large":
			return http.StatusRequestEntityTooLarge, err
		default:
			return http.StatusInternalServerError, fmt.Errorf("failed to decode json %w", err)
		}
	}
	if d.More() {
		return http.StatusBadRequest, fmt.Errorf("body must contain only one JSON object")
	}

	return http.StatusOK, nil
}
