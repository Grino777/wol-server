package server

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/Grino777/wol-server/internal/server/api"
	"github.com/Grino777/wol-server/internal/services/wol"
	"go.uber.org/zap"
)

type ServerImpl interface {
	Start() error
	Stop() error
}

type wolServer struct {
	ctx     context.Context
	server  *http.Server
	service wol.WOLService
}

type BootstrapOptions struct {
	Logger     *zap.SugaredLogger
	WOLService wol.WOLService
}

func Bootstrap(ctx context.Context, bo BootstrapOptions) ServerImpl {
	return NewServer(ctx, bo)
}

func NewServer(ctx context.Context, bo BootstrapOptions) ServerImpl {
	config := &tls.Config{}

	api := api.NewAPI(ctx, bo.WOLService)
	handler := api.Handler()

	server := &http.Server{
		Addr:      ":8080",
		Handler:   handler,
		TLSConfig: config,
	}

	return &wolServer{
		ctx:     ctx,
		server:  server,
		service: bo.WOLService,
	}
}

func (s *wolServer) Start() error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *wolServer) Stop() error {
	if err := s.server.Shutdown(context.Background()); err != nil {
		return err
	}
	return nil
}
