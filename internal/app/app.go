// Пакет app отвечает за инициализацию и управление жизненным циклом приложения,
// включая настройку базы данных, запуск сервера и обработку ошибок.
// Обеспечивает централизованное место для управления всеми компонентами приложения
// и их взаимодействием.
package app

import (
	"context"

	sqlite "github.com/Grino777/wol-server/internal/database/sqlite"
	"github.com/Grino777/wol-server/internal/database/sqlite/repo"
	"github.com/Grino777/wol-server/internal/server"
	service "github.com/Grino777/wol-server/internal/services/wol"
	"github.com/Grino777/wol-server/pkg/logging"
)

type AppImpl interface {
	Start() error
	Stop() error
}

type app struct {
	ctx    context.Context
	server server.ServerImpl
	db     sqlite.DatabaseImpl
}

func NewApp(ctx context.Context) (AppImpl, error) {
	db, err := sqlite.NewDatabase(ctx)
	if err != nil {
		logging.FromContext(ctx).Errorw("failed to create database", "error", err)
		return nil, err
	}

	migrator, err := sqlite.NewMigrator(db.Path())
	if err != nil {
		logging.FromContext(ctx).Errorw("failed to create migrator", "error", err)
		_ = db.Close()
		return nil, err
	}

	if err := migrator.Latest(); err != nil {
		logging.FromContext(ctx).Errorw("failed to run migrations", "error", err)
		_ = db.Close()
		return nil, err
	}

	serverRepo := repo.NewServerRepository(db.Connection())
	userRepo := repo.NewUserRepository(db.Connection())

	wolService := service.NewWOLService(serverRepo, userRepo)

	httpServer := server.Bootstrap(ctx, server.BootstrapOptions{
		Logger:     logging.FromContext(ctx),
		WOLService: wolService,
	})

	return &app{
		ctx:    ctx,
		server: httpServer,
		db:     db,
	}, nil
}

func (a *app) Start() error {
	if err := a.server.Start(); err != nil {
		logging.FromContext(a.ctx).Errorw("failed to start server", "error", err)
		return err
	}
	defer func() {
		if err := a.Stop(); err != nil {
			logging.FromContext(a.ctx).Errorw("failed to stop server", "error", err)
		}
	}()
	return nil
}

func (a *app) Stop() error {
	if err := a.server.Stop(); err != nil {
		logging.FromContext(a.ctx).Errorw("failed to stop server", "error", err)
		return err
	}

	if err := a.db.Close(); err != nil {
		logging.FromContext(a.ctx).Errorw("failed to close database", "error", err)
		return err
	}

	return nil
}
