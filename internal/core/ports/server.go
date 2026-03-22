package ports

import (
	"context"

	"github.com/Grino777/wol-server/internal/core/entity"
)

// ServerDB определяет интерфейс для работы с данными серверов в базе данных,
// включая методы для получения всех серверов, получения сервера по ID,
// создания нового сервера, обновления существующего сервера и удаления сервера по ID.
type ServerDB interface {
	FindAll(ctx context.Context) ([]*entity.Server, error)
	FindByID(ctx context.Context, id int) (*entity.Server, error)
	Create(ctx context.Context, server *entity.Server) error
	Update(ctx context.Context, server *entity.Server) error
	Delete(ctx context.Context, id int) error
}
