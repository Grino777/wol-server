// Пакет sqlite реализует базу данных SQLite для приложения WOL Server.
// Предоставляет реализацию интерфейса database, определенного в пакете database,
// и использует драйвер go-sqlite3 для взаимодействия с базой данных.
// Пакет также обеспечивает проверку наличия базы данных
// и создание необходимых директорий при запуске приложения.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Grino777/wol-server/pkg/logging"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DbPath = "/var/lib/wol-server"
	DbName = "wol.db"
)

type DatabaseImpl interface {
	Connect() error
	Connection() *sql.DB
	Path() string
	Close() error
}

type sqlDatabase struct {
	ctx      context.Context
	sqliteDB *sql.DB
	dbPath   string
}

func NewDatabase(ctx context.Context) (DatabaseImpl, error) {
	dbPath := fmt.Sprintf("%s%s/%s", os.Getenv("HOME"), DbPath, DbName)

	if err := checkDB(dbPath, ctx); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logging.FromContext(ctx).Errorw("failed to open db", "error", err)
		return nil, err
	}
	if err := db.Ping(); err != nil {
		logging.FromContext(ctx).Errorw("failed to ping db", "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(1)

	return &sqlDatabase{
		ctx:      ctx,
		dbPath:   dbPath,
		sqliteDB: db,
	}, nil
}

func (d *sqlDatabase) Connect() error {
	return nil
}

func (d *sqlDatabase) Connection() *sql.DB {
	return d.sqliteDB
}

func (d *sqlDatabase) Path() string {
	return d.dbPath
}

func (d *sqlDatabase) Close() error {
	if err := d.sqliteDB.Close(); err != nil {
		logging.FromContext(d.ctx).Errorw("failed to close database", "error", err)
		return err
	}
	return nil
}

func checkDB(dbPath string, ctx context.Context) error {
	if _, err := os.Stat(dbPath); err != nil {
		logging.FromContext(ctx).Warn("database file does not exist")

		if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
			logging.FromContext(ctx).Errorw("failed to create database directory", "error", err)
			return err
		}
		logging.FromContext(ctx).Info("database directory created")
	}

	return nil
}
