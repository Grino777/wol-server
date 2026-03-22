package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Grino777/wol-server/internal/core/entity"
	"github.com/Grino777/wol-server/internal/core/ports"
	sqlite3 "github.com/mattn/go-sqlite3"
)

type serverRepository struct {
	db *sql.DB
}

// compile check
var _ ports.ServerDB = (*serverRepository)(nil)

// NewServerRepository создает новый экземпляр ServerRepository,
// который реализует интерфейс database.ServerDB,
// и использует переданное соединение с базой данных для выполнения операций CRUD над сущностью Server.
func NewServerRepository(db *sql.DB) ports.ServerDB {
	return &serverRepository{db: db}
}

func (r *serverRepository) FindAll(ctx context.Context) ([]*entity.Server, error) {
	query := `
		SELECT 
			id, name, mac_address, ip_address, port, created_at, updated_at
		FROM 
			servers
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}
	defer rows.Close()

	servers, err := r.scanServers(rows)
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func (r *serverRepository) FindByID(ctx context.Context, id int) (*entity.Server, error) {
	query := `
		SELECT
			id, name, mac_address, ip_address, port, created_at, updated_at
		FROM 
			servers
		WHERE
			id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	server, err := r.scanServer(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return server, nil
}

func (r *serverRepository) Create(ctx context.Context, s *entity.Server) error {
	query := `
		INSERT INTO servers 
			(name, mac_address, ip_address, port, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := r.db.ExecContext(ctx, query, s.Name, s.MacAddress, s.IpAddress, s.Port)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && (sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique || sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey) {
			return entity.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (r *serverRepository) Update(ctx context.Context, s *entity.Server) error {
	query := `
		UPDARE servers
		SET
			name = $2
			mac_address = $3
			ip_address = $4
			updated_at = $5
		WHERE
			id = $1
	`

	if _, err := r.db.ExecContext(ctx, query, s.Id, s.MacAddress, s.IpAddress, s.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *serverRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM servers
		WHERE 
			id = $1
	`

	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *serverRepository) scanServers(rows *sql.Rows) ([]*entity.Server, error) {
	servers := make([]*entity.Server, 0)

	for rows.Next() {
		server, err := r.scanServer(rows)
		if err != nil {
			return nil, err
		}

		servers = append(servers, server)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func (r *serverRepository) scanServer(row scanner) (*entity.Server, error) {
	var m entity.Server

	if err := row.Scan(
		&m.Id,
		&m.Name,
		&m.MacAddress,
		&m.IpAddress,
		&m.Port,
		&m.CreatedAt,
		&m.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &m, nil
}
