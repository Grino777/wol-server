// Пакет database определяет интерфейсы и типы для операций с базой данных,
// включая управление соединениями, миграции,
// и репозитории для серверов и пользователей.
// Он служит уровнем абстракции для базовой реализации базы данных,
// обеспечивая гибкость и разделение задач в архитектуре приложения.
package database

import (
	"database/sql"

	"github.com/Grino777/wol-server/internal/core/ports"
)

// Интерфейс для управления соединениями с базой данных,
// который может быть реализован для различных СУБД,
// таких как SQLite, PostgreSQL и т.д.
type DB interface {
	Connection() *sql.DB
	Close() error
}

// Интерфейс для управления миграциями базы данных,
// который может быть реализован для различных СУБД,
// таких как SQLite, PostgreSQL и т.д.
type Migrator interface {
	Latest() error
	First() error
	UpTo(version uint) error
	DownTo(version uint) error
}

// Репозитории для серверов и пользователей,
// которые реализуют интерфейсы из ports
type (
	ServerRepository = ports.ServerDB
	UserRepository   = ports.UserDB
)
