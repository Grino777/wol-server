package usecase

import (
	"context"

	"github.com/Grino777/wol-server/internal/core/entity"
)

type UserService interface {
	GetUsers(ctx context.Context) ([]entity.User, error)
	GetUserByID(ctx context.Context, id int) (entity.User, error)
	// CreateUser(ctx context.Context, user entity.User) error
	UpdateUser(ctx context.Context, user entity.User) error
	DeleteUser(ctx context.Context, id int) error
}
