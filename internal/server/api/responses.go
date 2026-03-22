package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Grino777/wol-server/internal/core/entity"
	v1 "github.com/Grino777/wol-server/pkg/api/v1"
	"github.com/Grino777/wol-server/pkg/jsonutil"
)

// decode is function to unmarshalling request into data.
//
// The request must contain the following headers:
//   - Content-Type: application/json
//
// The function returns two values: an http status code and an error containing
// information prepared for a response. If status 500 (Internal Server Error)
// is returned, the error will be logged in function.
//
// HTTP statuses that can be returned:
//   - 200: data was successfully decoded
//   - 400: invalid request body
//   - 413: request body too large
//   - 415: content-type is not application/json
//   - 500: decode failed
func (api *API) decode(w http.ResponseWriter, r *http.Request, data any) (status int, err error) {
	status, err = jsonutil.UnmarshalRequest(w, r, data) // 200, 400, 413, 415, 500
	if err != nil {
		if status == http.StatusInternalServerError {
			api.logger.Errorf("json request unmarshaling error: %s", err)
		}
		return status, err
	}
	return status, nil
}

func (e ErrorMessage) Error() string {
	return e.Message
}

// respond is responding function that marshal data in provided response
// content type and write it to http.ResponseWriter (by default application/json).
// This should be called just once per http handler.
//
// This function also defines the general response headers.
// Provided headers:
//   - Content-Type: application/json
//
// If data implements a Public interface, it will call and override the original data.
// It is useful for hidden replacing of request without additional logic.
func (api *API) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	if obj, ok := data.(v1.Public); ok {
		data = obj.Public()
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		api.logger.Errorf("json response marshaling error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.setHeader(w, r)

	w.WriteHeader(status)
	if _, err := io.Copy(w, &buf); err != nil {
		api.logger.Errorf("response writing error: %s", err)
	}
}

func (api *API) respondWithNotFound(w http.ResponseWriter, r *http.Request) {
	resp := ErrorMessage{
		Code:    v1.CodeNotFound,
		Message: entity.ErrNotFound.Error(),
	}
	api.respond(w, r, &resp, http.StatusNotFound)
}

func (api *API) respondWithError(w http.ResponseWriter, r *http.Request, code, msg string, status int) {
	resp := ErrorMessage{
		Code:    code,
		Message: msg,
	}
	api.respond(w, r, &resp, status)
}

// TODO:
func (api *API) respondWithAPIErrorResponse(w http.ResponseWriter, r *http.Request, funcName string, err error) {
	switch {
	case errors.Is(err, entity.ErrAlreadyExists):
		api.respondWithError(w, r, v1.CodeConflict, v1.MessageAlreadyExist, http.StatusConflict)
		return
	case errors.Is(err, entity.ErrNotFound):
		api.respondWithError(w, r, v1.CodeNotFound, v1.MessageNotFound, http.StatusBadRequest)
		return
	default:
		api.logger.Errorf("failed to handle request in `%s`: %v", funcName, err)
		api.respondWithInternalServerError(w, r)
		return
	}
}

func (api *API) respondWithInternalServerError(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, entity.ErrInternalServerError.Error(), http.StatusInternalServerError)
}

func (api *API) respondNotImplemented(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (api *API) respondWithBadRequest(w http.ResponseWriter, r *http.Request) {
	api.respondWithError(w, r, v1.CodeBadRequest, v1.CodeInvalidData, http.StatusBadRequest)
}

// decodeOrRespondWithError is a helper for handling request decoding error.
// In case any decoding error is triggered, the function will respond with
// an error-specific response.
func (api *API) decodeOrRespondWithError(w http.ResponseWriter, r *http.Request, data interface{}) (success bool) {
	status, err := api.decode(w, r, data)
	if err == nil {
		return true
	}

	// internal error
	if status == http.StatusInternalServerError { // 500
		api.respondWithInternalServerError(w, r)
		return false
	}

	// client error
	api.respondWithError(w, r,
		v1.CodeBadRequest, // code
		err.Error(),       // message
		status,            // status
	)
	return false
}

func (api *API) respondSuccess(w http.ResponseWriter, r *http.Request, message v1.SuccessfulMessage) {
	api.respond(w, r, v1.SuccessfulResponse{Status: string(message)}, http.StatusOK)
}
