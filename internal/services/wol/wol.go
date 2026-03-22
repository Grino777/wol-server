package wol

import (
	"github.com/Grino777/wol-server/internal/core/ports"
	"github.com/Grino777/wol-server/internal/core/usecase"
)

type WOLRequest struct {
	ID int
}

type WOLService interface {
	ServersService() usecase.ServerService
	UsersService() usecase.UserService
	ActionsService() usecase.ActionsService
}

type wolService struct {
	serverRepository ports.ServerDB
	userRepository   ports.UserDB
}

// compile check
var _ WOLService = (*wolService)(nil)

func NewWOLService(
	serverRepository ports.ServerDB,
	userRepository ports.UserDB,
) WOLService {
	return &wolService{
		serverRepository: serverRepository,
		userRepository:   userRepository,
	}
}

func (s *wolService) ServersService() usecase.ServerService {
	return s
}

func (s *wolService) UsersService() usecase.UserService {
	return s
}

func (s *wolService) ActionsService() usecase.ActionsService {
	return s
}
