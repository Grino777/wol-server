package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestMigratorLatestDownloadsAppliesAndCleansUp(t *testing.T) {
	t.Parallel()

	migrationFiles := map[string]string{
		"000001_init.up.sql": `
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);`,
		"000001_init.down.sql": `
DROP TABLE users;`,
	}

	apiURL, httpClient := newGitHubMigrationsMockServer(t, migrationFiles)

	dbPath := filepath.Join(t.TempDir(), "wol.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	tempRootDir := filepath.Join(t.TempDir(), "tmp", "wol-server")
	migrator := &Migrator{
		dbPath:           dbPath,
		httpClient:       httpClient,
		migrationsAPIURL: apiURL,
		tempRootDir:      tempRootDir,
	}

	if err := migrator.Latest(); err != nil {
		t.Fatalf("apply latest migrations: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO users (name) VALUES ('alice')`); err != nil {
		t.Fatalf("users table was not created by migrations: %v", err)
	}

	if _, err := os.Stat(tempRootDir); !os.IsNotExist(err) {
		t.Fatalf("temporary migrations directory must be removed, stat error: %v", err)
	}
}

func TestMigratorLatestCleansUpWhenMigrationFails(t *testing.T) {
	t.Parallel()

	migrationFiles := map[string]string{
		"000001_broken.up.sql": `
CREATE TABLE broken (
    id INTEGER PRIMARY KEY
;
`,
		"000001_broken.down.sql": `
DROP TABLE broken;`,
	}

	apiURL, httpClient := newGitHubMigrationsMockServer(t, migrationFiles)

	dbPath := filepath.Join(t.TempDir(), "wol.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	tempRootDir := filepath.Join(t.TempDir(), "tmp", "wol-server")
	migrator := &Migrator{
		dbPath:           dbPath,
		httpClient:       httpClient,
		migrationsAPIURL: apiURL,
		tempRootDir:      tempRootDir,
	}

	err = migrator.Latest()
	if err == nil {
		t.Fatal("expected migration error, got nil")
	}

	if _, statErr := os.Stat(tempRootDir); !os.IsNotExist(statErr) {
		t.Fatalf("temporary migrations directory must be removed on error, stat error: %v", statErr)
	}
}

func TestMigratorLatestCleansUpWhenFetchFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "temporary github failure", http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	dbPath := filepath.Join(t.TempDir(), "wol.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	tempRootDir := filepath.Join(t.TempDir(), "tmp", "wol-server")
	migrator := &Migrator{
		dbPath:           dbPath,
		httpClient:       server.Client(),
		migrationsAPIURL: server.URL + "/api/migrations",
		tempRootDir:      tempRootDir,
	}

	err = migrator.Latest()
	if err == nil {
		t.Fatal("expected migrations list fetch error, got nil")
	}

	if _, statErr := os.Stat(tempRootDir); !os.IsNotExist(statErr) {
		t.Fatalf("temporary migrations directory must be removed on fetch error, stat error: %v", statErr)
	}
}

func TestMigratorLatestRecoversDirtyMigrationState(t *testing.T) {
	t.Parallel()

	migrationFiles := map[string]string{
		"000001_users.up.sql": `
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);`,
		"000001_users.down.sql": `
DROP TABLE users;`,
		"000002_broken.up.sql": `
CREATE TABLE broken (
    id INTEGER PRIMARY KEY
;
`,
		"000002_broken.down.sql": `
DROP TABLE IF EXISTS broken;`,
	}

	apiURL, httpClient := newGitHubMigrationsMockServer(t, migrationFiles)

	dbPath := filepath.Join(t.TempDir(), "wol.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	tempRootDir := filepath.Join(t.TempDir(), "tmp", "wol-server")
	migrator := &Migrator{
		dbPath:           dbPath,
		httpClient:       httpClient,
		migrationsAPIURL: apiURL,
		tempRootDir:      tempRootDir,
	}

	err = migrator.Latest()
	if err == nil {
		t.Fatal("expected migration error, got nil")
	}

	version, dirty, stateErr := readSchemaMigrationState(db)
	if stateErr != nil {
		t.Fatalf("read schema migration state: %v", stateErr)
	}
	if version != 1 {
		t.Fatalf("expected recovered version 1, got %d", version)
	}
	if dirty {
		t.Fatal("expected dirty=false after recovery")
	}

	if _, statErr := os.Stat(tempRootDir); !os.IsNotExist(statErr) {
		t.Fatalf("temporary migrations directory must be removed on recovery path, stat error: %v", statErr)
	}
}

func readSchemaMigrationState(db *sql.DB) (uint, bool, error) {
	var version uint
	var dirty bool
	if err := db.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&version, &dirty); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return version, dirty, nil
}

func newGitHubMigrationsMockServer(t *testing.T, files map[string]string) (string, *http.Client) {
	t.Helper()

	fileNames := make([]string, 0, len(files))
	for fileName := range files {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case request.URL.Path == "/api/migrations":
			response := make([]githubMigrationFile, 0, len(fileNames))
			for _, fileName := range fileNames {
				response = append(response, githubMigrationFile{
					Name:        fileName,
					Type:        "file",
					DownloadURL: "http://" + request.Host + "/files/" + fileName,
				})
			}

			writer.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(writer).Encode(response); err != nil {
				t.Fatalf("encode mock response: %v", err)
			}
		case strings.HasPrefix(request.URL.Path, "/files/"):
			fileName := path.Base(request.URL.Path)
			content, ok := files[fileName]
			if !ok {
				http.NotFound(writer, request)
				return
			}
			if _, err := writer.Write([]byte(content)); err != nil {
				t.Fatalf("write file content for %q: %v", fileName, err)
			}
		default:
			http.NotFound(writer, request)
		}
	})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return server.URL + "/api/migrations", server.Client()
}
