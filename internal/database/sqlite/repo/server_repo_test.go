package repo

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/Grino777/wol-server/internal/core/entity"
	_ "github.com/mattn/go-sqlite3"
)

func TestServerRepositoryCreateReturnsAlreadyExistsOnUniqueViolation(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite3", "file:repo-create-unique?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	schema := `
		CREATE TABLE servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			mac_address TEXT NOT NULL UNIQUE,
			ip_address TEXT NOT NULL UNIQUE,
			port INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	repository := NewServerRepository(db)
	ctx := context.Background()

	firstServer := &entity.Server{
		Id:         1,
		Name:       "primary",
		MacAddress: "AA:BB:CC:DD:EE:FF",
		IpAddress:  "192.168.0.10",
		Port:       9,
	}
	if err := repository.Create(ctx, firstServer); err != nil {
		t.Fatalf("create initial server: %v", err)
	}

	duplicateServer := &entity.Server{
		Id:         2,
		Name:       "secondary",
		MacAddress: firstServer.MacAddress,
		IpAddress:  "192.168.0.11",
		Port:       9,
	}
	err = repository.Create(ctx, duplicateServer)
	if !errors.Is(err, entity.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got: %v", err)
	}
}
