package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *API) handlerOff(w http.ResponseWriter, r *http.Request) {
	const op = "handlerOff"

	ctx := r.Context()

	serverID, err := parseIntPathParam(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := api.wakeService.OffServer(ctx, serverID); err != nil {
		api.respondWithAPIErrorResponse(w, r, op, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) handlerWake(w http.ResponseWriter, r *http.Request) {
	const op = "handlerWake"

	ctx := r.Context()

	serverID, err := parseIntPathParam(chi.URLParam(r, "id"))
	if err != nil {
		api.respondWithBadRequest(w, r)
		return
	}

	if err := api.wakeService.WakeServer(ctx, serverID); err != nil {
		api.respondWithAPIErrorResponse(w, r, op, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
