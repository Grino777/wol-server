package ports

import (
	"context"

	"github.com/Grino777/wol-server/internal/core/entity"
)

// UserDB определяет интерфейс для работы с данными пользователей в базе данных,
// включая методы для получения всех пользователей, получения пользователя по ID,
// создания нового пользователя, обновления существующего пользователя и удаления пользователя по ID.
type UserDB interface {
	FindAll(ctx context.Context) ([]*entity.User, error)
	FindByID(ctx context.Context, id int) (*entity.User, error)
	// Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int) error
}
