// Пакет main является точкой входа в приложение,
// инициализирует контекст с логгером и запускает основное приложение.
package main

import (
	"context"

	"github.com/Grino777/wol-server/internal/app"
	"github.com/Grino777/wol-server/pkg/logging"
)

func main() {
	ctx := logging.WithLogger(context.Background(), logging.NewLoggerFromEnv())

	app, err := app.NewApp(ctx)
	if err != nil {
		logging.FromContext(ctx).Errorw("failed to create app", "error", err)
		return
	}
	if err := app.Start(); err != nil {
		logging.FromContext(ctx).Errorw("failed to start app", "error", err)
		return
	}
}
