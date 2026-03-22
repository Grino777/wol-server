package wol

import (
	"context"

	"github.com/Grino777/wol-server/internal/core/entity"
)

func (s *wolService) GetServers(ctx context.Context) ([]*entity.Server, error) {
	return s.serverRepository.FindAll(ctx)
}

func (s *wolService) GetServerByID(ctx context.Context, id int) (*entity.Server, error) {
	return s.serverRepository.FindByID(ctx, id)
}

func (s *wolService) CreateServer(ctx context.Context, server *entity.Server) error {
	return s.serverRepository.Create(ctx, server)
}

func (s *wolService) UpdateServer(ctx context.Context, server *entity.Server) error {
	return s.serverRepository.Update(ctx, server)
}

func (s *wolService) DeleteServer(ctx context.Context, id int) error {
	return s.serverRepository.Delete(ctx, id)
}
