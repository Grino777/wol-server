package api

import (
	"net/http"
	"strconv"
)

type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (api *API) setHeader(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}

func parseIntPathParam(raw string) (int, error) {
	return strconv.Atoi(raw)
}
