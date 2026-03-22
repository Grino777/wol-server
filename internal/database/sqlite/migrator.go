// Пакет sqlite реализует мигратор для базы данных SQLite,
// который загружает SQL-скрипты миграций из репозитория на GitHub и применяет их к базе данных.
// Использует библиотеку golang-migrate для управления миграциями
// и обеспечивает восстановление из "dirty" состояния в случае ошибок.
// Миграционные файлы загружаются во временную директорию, которая удаляется после применения миграций.
package sqlite

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Grino777/wol-server/internal/database"
	"github.com/Grino777/wol-server/pkg/logging"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	"go.uber.org/zap"
)

const (
	githubMigrationsAPIURL = "https://api.github.com/repos/Grino777/wol-server/contents/migrations/sqlite?ref=main"
	tempMigrationsSubDir   = "migrations"
	maxMigrationFileSize   = 4 * 1024 * 1024
)

type Migrator struct {
	dbPath           string
	httpClient       *http.Client
	migrationsAPIURL string
	tempRootDir      string
	logger           *zap.SugaredLogger
}

func NewMigrator(dbPath string) (database.Migrator, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	return &Migrator{
		dbPath: dbPath,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		migrationsAPIURL: githubMigrationsAPIURL,
		tempRootDir:      filepath.Join(homeDir, "tmp", "wol-server"),
		logger:           logging.DefaultLogger(),
	}, nil
}

func (m *Migrator) Latest() (err error) {
	return m.runWithMigrator(func(migrator *migrate.Migrate) error {
		migrateErr := migrator.Up()
		if migrateErr != nil && !errors.Is(migrateErr, migrate.ErrNoChange) {
			return migrateErr
		}
		return nil
	})
}

func (m *Migrator) First() error {
	return m.UpTo(1)
}

func (m *Migrator) UpTo(version uint) error {
	return m.runWithMigrator(func(migrator *migrate.Migrate) error {
		migrateErr := migrator.Migrate(version)
		if migrateErr != nil && !errors.Is(migrateErr, migrate.ErrNoChange) {
			return migrateErr
		}
		return nil
	})
}

func (m *Migrator) DownTo(version uint) error {
	return m.runWithMigrator(func(migrator *migrate.Migrate) error {
		migrateErr := migrator.Migrate(version)
		if migrateErr != nil && !errors.Is(migrateErr, migrate.ErrNoChange) {
			return migrateErr
		}
		return nil
	})
}

type githubMigrationFile struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

func (m *Migrator) runWithMigrator(run func(*migrate.Migrate) error) (err error) {
	migrationsDir, cleanup, err := m.prepareMigrationsDir()
	if err != nil {
		if cleanup != nil {
			if cleanupErr := cleanup(); cleanupErr != nil {
				return fmt.Errorf("prepare migrations directory: %w", errors.Join(err, fmt.Errorf("cleanup temporary migrations directory: %w", cleanupErr)))
			}
		}
		return fmt.Errorf("prepare migrations directory: %w", err)
	}
	defer func() {
		if cleanupErr := cleanup(); cleanupErr != nil {
			err = joinErrors(err, fmt.Errorf("cleanup temporary migrations directory: %w", cleanupErr))
		}
	}()

	migrator, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsDir),
		fmt.Sprintf("sqlite3://%s", m.dbPath),
	)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		sourceErr, databaseErr := migrator.Close()
		closeErr := errors.Join(sourceErr, databaseErr)
		if closeErr != nil {
			err = joinErrors(err, fmt.Errorf("close migrator: %w", closeErr))
		}
	}()

	if runErr := run(migrator); runErr != nil {
		return joinErrors(runErr, m.recoverDirtyState(migrator, runErr))
	}

	return nil
}

func (m *Migrator) recoverDirtyState(migrator *migrate.Migrate, runErr error) error {
	version, dirty, err := migrator.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return nil
		}
		return fmt.Errorf("read migration version after failure: %w", err)
	}
	if !dirty {
		return nil
	}

	logger := m.logger
	logger.Errorw(
		"migration failed in dirty state, attempting recovery",
		"error", runErr,
		"version", version,
		"dirty", dirty,
	)

	if err := migrator.Force(int(version)); err != nil {
		return fmt.Errorf("force migration version %d before recovery: %w", version, err)
	}

	if version > 0 {
		if err := migrator.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("rollback migration from version %d to %d: %w", version, version-1, err)
		}
	}

	currentVersion, currentDirty, err := migrator.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			logger.Errorw(
				"dirty migration recovered",
				"error", runErr,
				"failed_version", version,
				"failed_dirty", dirty,
				"current_version", 0,
				"current_dirty", false,
			)
			return nil
		}
		return fmt.Errorf("read migration version after recovery: %w", err)
	}

	logger.Errorw(
		"dirty migration recovered",
		"error", runErr,
		"failed_version", version,
		"failed_dirty", dirty,
		"current_version", currentVersion,
		"current_dirty", currentDirty,
	)
	if currentDirty {
		return fmt.Errorf("dirty flag is still true after recovery at version %d", currentVersion)
	}

	return nil
}

func (m *Migrator) prepareMigrationsDir() (string, func() error, error) {
	if err := os.RemoveAll(m.tempRootDir); err != nil {
		return "", nil, fmt.Errorf("remove stale temporary root directory: %w", err)
	}

	migrationsDir := filepath.Join(m.tempRootDir, tempMigrationsSubDir)
	if err := os.MkdirAll(migrationsDir, 0700); err != nil {
		return "", nil, fmt.Errorf("create temporary migrations directory: %w", err)
	}

	cleanup := func() error {
		return os.RemoveAll(m.tempRootDir)
	}

	files, err := m.fetchMigrationFiles()
	if err != nil {
		return "", cleanup, err
	}

	for _, file := range files {
		if err := m.downloadMigrationFile(migrationsDir, file); err != nil {
			return "", cleanup, err
		}
	}

	return migrationsDir, cleanup, nil
}

func (m *Migrator) fetchMigrationFiles() ([]githubMigrationFile, error) {
	request, err := http.NewRequest(http.MethodGet, m.migrationsAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create github request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")

	response, err := m.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch migrations list from github: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("github migrations list request failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var content []githubMigrationFile
	if err := json.NewDecoder(response.Body).Decode(&content); err != nil {
		return nil, fmt.Errorf("decode github migrations list response: %w", err)
	}

	migrationFiles := make([]githubMigrationFile, 0, len(content))
	for _, file := range content {
		if file.Type != "file" {
			continue
		}
		if !strings.HasSuffix(file.Name, ".sql") {
			continue
		}
		if file.DownloadURL == "" {
			return nil, fmt.Errorf("missing download URL for migration file %q", file.Name)
		}
		migrationFiles = append(migrationFiles, file)
	}
	if len(migrationFiles) == 0 {
		return nil, errors.New("no migration files found in github response")
	}

	sort.Slice(migrationFiles, func(i, j int) bool {
		return migrationFiles[i].Name < migrationFiles[j].Name
	})

	m.logger.Debugf("downloads %d migrations files", len(migrationFiles))

	return migrationFiles, nil
}

func (m *Migrator) downloadMigrationFile(migrationsDir string, file githubMigrationFile) error {
	if path.Base(file.Name) != file.Name {
		return fmt.Errorf("invalid migration file name %q", file.Name)
	}

	request, err := http.NewRequest(http.MethodGet, file.DownloadURL, nil)
	if err != nil {
		return fmt.Errorf("create request for migration file %q: %w", file.Name, err)
	}

	response, err := m.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("download migration file %q: %w", file.Name, err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("download migration file %q failed with status %d: %s", file.Name, response.StatusCode, strings.TrimSpace(string(body)))
	}

	destinationPath := filepath.Join(migrationsDir, file.Name)
	destinationFile, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open destination for migration file %q: %w", file.Name, err)
	}
	defer func() {
		_ = destinationFile.Close()
	}()

	written, err := io.Copy(destinationFile, io.LimitReader(response.Body, maxMigrationFileSize+1))
	if err != nil {
		return fmt.Errorf("write migration file %q: %w", file.Name, err)
	}
	if written > maxMigrationFileSize {
		_ = os.Remove(destinationPath)
		return fmt.Errorf("migration file %q exceeds maximum allowed size of %d bytes", file.Name, maxMigrationFileSize)
	}

	return nil
}

func joinErrors(baseErr, extraErr error) error {
	if extraErr == nil {
		return baseErr
	}
	if baseErr == nil {
		return extraErr
	}
	return errors.Join(baseErr, extraErr)
}
