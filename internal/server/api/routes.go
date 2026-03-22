package api

import (
	"context"

	"github.com/Grino777/wol-server/internal/middleware"
	"github.com/Grino777/wol-server/pkg/logging"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func (api *API) SetupRoutes(ctx context.Context) chi.Router {
	r := api.router
	if r == nil {
		r = chi.NewRouter()
		api.router = r
	}

	r.Use(middleware.ContextLogger(logging.FromContext(ctx)))
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(middleware.AuthMiddleware())
	r.Use(middleware.MaxBytes())
	r.Use(middleware.RequestLogger(logging.FromContext(ctx)))
	r.Use(chimw.StripSlashes)

	r.Route("/api/v1", func(r chi.Router) {

		// ===== server collection routes =====
		r.Route("/servers", func(r chi.Router) {
			r.Get("/", api.handlerGetServers)
			r.Get("/{id}", api.handlerGetServer)
			r.Post("/", api.handlerCreateServer)
			r.Delete("/{id}", api.handlerDeleteServer)
			r.Patch("/{id}", api.handlerUpdateServer)
		})

		// ===== user collection routes =====
		r.Route("/users", func(r chi.Router) {
			r.Get("/", api.handlerGetUsers)
			r.Get("/{id}", api.handlerGetUser)
			r.Post("/{id}", api.handlerCreateUser)
			r.Delete("/{id}", api.handlerDeleteUser)
			r.Patch("/{id}", api.handlerUpdateUser)
		})

		// ===== action routes =====
		r.Route("/wake", func(r chi.Router) {
			r.Post("/{id}", api.handlerWake)
		})

		r.Route("/off", func(r chi.Router) {
			r.Post("/{id}", api.handlerOff)
		})
	})

	return r
}
