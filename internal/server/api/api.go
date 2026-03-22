package api

import (
	"context"
	"net/http"

	"github.com/Grino777/wol-server/internal/core/usecase"
	"github.com/Grino777/wol-server/internal/services/wol"
	"github.com/Grino777/wol-server/pkg/logging"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type API struct {
	router chi.Router
	logger *zap.SugaredLogger

	userService   usecase.UserService
	serverService usecase.ServerService
	wakeService   usecase.ActionsService
}

func NewAPI(ctx context.Context, wol wol.WOLService) *API {
	a := &API{
		router:        chi.NewRouter(),
		logger:        logging.FromContext(ctx),
		userService:   wol.UsersService(),
		serverService: wol.ServersService(),
		wakeService:   wol.ActionsService(),
	}
	a.SetupRoutes(ctx)
	return a
}

func (api *API) Handler() http.Handler {
	if api == nil {
		return http.NotFoundHandler()
	}
	return api.router
}
