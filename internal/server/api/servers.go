package api

import (
	"net/http"

	"github.com/Grino777/wol-server/internal/core/entity"
	v1 "github.com/Grino777/wol-server/pkg/api/v1"
	"github.com/go-chi/chi/v5"
)

func (api *API) handlerGetServers(w http.ResponseWriter, r *http.Request) {
	const op = "handlerGetServers"

	ctx := r.Context()

	servers, err := api.serverService.GetServers(ctx)
	if err != nil {
		api.respondWithAPIErrorResponse(w, r, op, err)
		return
	}

	api.respond(w, r, servers, http.StatusOK)
}

func (api *API) handlerGetServer(w http.ResponseWriter, r *http.Request) {
	const op = "handlerGetServer"

	ctx := r.Context()

	id, err := parseIntPathParam(chi.URLParam(r, "id"))
	if err != nil {
		api.respondWithError(w, r, v1.CodeBadRequest, v1.CodeInvalidData, http.StatusBadRequest)
		return
	}

	server, err := api.serverService.GetServerByID(ctx, id)
	if err != nil {
		api.respondWithAPIErrorResponse(w, r, op, err)
		return
	}

	api.respond(w, r, server, http.StatusOK)
}

func (api *API) handlerCreateServer(w http.ResponseWriter, r *http.Request) {
	const op = "handlerCreateServer"

	ctx := r.Context()

	var server entity.Server
	if ok := api.decodeOrRespondWithError(w, r, &server); !ok {
		return
	}

	if err := api.serverService.CreateServer(ctx, &server); err != nil {
		api.respondWithAPIErrorResponse(w, r, op, err)
		return
	}

	api.respondSuccess(w, r, v1.Created)
}

func (api *API) handlerDeleteServer(w http.ResponseWriter, r *http.Request) {
	const op = "handlerDeleteServer"

	ctx := r.Context()

	id, err := parseIntPathParam(chi.URLParam(r, "id"))
	if err != nil {
		api.respondWithAPIErrorResponse(w, r, op, err)
		return
	}

	if err := api.serverService.DeleteServer(ctx, id); err != nil {
		api.respondWithAPIErrorResponse(w, r, op, err)
		return
	}

	api.respondSuccess(w, r, v1.Deleted)
}

func (api *API) handlerUpdateServer(w http.ResponseWriter, r *http.Request) {
	const op = "handlerUpdateServer"

	api.respondNotImplemented(w, r)
}
