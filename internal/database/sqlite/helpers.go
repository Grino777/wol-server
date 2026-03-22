package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/Grino777/wol-server/pkg/logging"
)

func (d *sqlDatabase) ExecContextWithTimer(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := d.Connection().ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		logging.FromContext(ctx).Errorw("failed to execute query", "query", query, "args", args, "duration", duration, "error", err)
	} else {
		logging.FromContext(ctx).Debugw("query executed successfully", "query", query, "args", args, "duration", duration)
	}

	return result, err
}
