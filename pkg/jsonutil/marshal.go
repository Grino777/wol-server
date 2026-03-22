package jsonutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Grino777/wol-server/pkg/logging"
)

// MarshalResponse is a helper function to write an object to the http.ResponseWriter
// with fallback error template.
func MarshalResponse(w http.ResponseWriter, r *http.Request, response interface{}, status int) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	b, err := json.Marshal(response)
	if err != nil {
		logging.FromContext(ctx).Errorf(
			"response marshaling error: %v", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
		msg := escapeJSON(err.Error())
		fmt.Fprintf(w, jsonErrTmpl, msg)
		return
	}

	w.WriteHeader(status)
	fmt.Fprintf(w, "%s", b)
}

// escapeJSON does primitive JSON escaping.
func escapeJSON(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// jsonErrTmpl is the template to use when returning a JSON error. It is
// rendered using Printf, not json.Encode, so values must be escaped by the
// caller.
const jsonErrTmpl = `{"error":"%s"}`
