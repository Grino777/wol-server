package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Grino777/wol-server/internal/core/entity"
	"github.com/Grino777/wol-server/internal/core/ports"
)

type userRepository struct {
	db *sql.DB
}

// compile check
var _ ports.UserDB = (*userRepository)(nil)

// NewUserRepository создает новый экземпляр UserRepository,
// который реализует интерфейс database.UserDB,
// и использует переданное соединение с базой данных для выполнения операций CRUD над сущностью User.
func NewUserRepository(db *sql.DB) ports.UserDB {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll(ctx context.Context) ([]*entity.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, username, password_hash, created_at, updated_at
		 FROM users
		 ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		user := &entity.User{}
		var id int64

		if err := rows.Scan(
			&id,
			&user.Username,
			&user.PasswordHash,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.Id = uint(id)
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*entity.User, error) {
	query := `
		SELECT 
			id, username, password_hash, created_at, updated_at
		FROM 
			users
		WHERE 
			id = ?
		LIMIT 1	
	`

	row := r.db.QueryRowContext(
		ctx,
		query,
		id,
	)

	user := &entity.User{}
	var userID int64
	if err := row.Scan(
		&userID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	user.Id = uint(userID)

	return user, nil
}

// func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
// 	query := `
// 		INSERT INTO users
// 			(username, password_hash)
// 		VALUES
// 			(?, ?)
// 	`
// 	result, err := r.db.ExecContext(
// 		ctx,
// 		query,
// 		user.Username,
// 		user.PasswordHash,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	id, err := result.LastInsertId()
// 	if err != nil {
// 		return err
// 	}
// 	user.Id = uint(id)

// 	return nil
// }

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET
			username = ?,
			password_hash = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.PasswordHash,
		user.Id,
	)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}
