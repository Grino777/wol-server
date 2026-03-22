package usecase

import (
	"context"

	"github.com/Grino777/wol-server/internal/core/entity"
)

type ServerService interface {
	GetServers(ctx context.Context) ([]*entity.Server, error)
	GetServerByID(ctx context.Context, id int) (*entity.Server, error)
	CreateServer(ctx context.Context, server *entity.Server) error
	UpdateServer(ctx context.Context, server *entity.Server) error
	DeleteServer(ctx context.Context, id int) error
}
