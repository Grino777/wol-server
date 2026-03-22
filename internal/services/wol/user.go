package wol

import (
	"context"

	"github.com/Grino777/wol-server/internal/core/entity"
)

func (s *wolService) GetUsers(ctx context.Context) ([]entity.User, error) {
	users, err := s.userRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]entity.User, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}
		result = append(result, *user)
	}

	return result, nil
}

func (s *wolService) GetUserByID(ctx context.Context, id int) (entity.User, error) {
	user, err := s.userRepository.FindByID(ctx, id)
	if err != nil {
		return entity.User{}, err
	}
	if user == nil {
		return entity.User{}, nil
	}

	return *user, nil
}

// func (s *wolService) CreateUser(ctx context.Context, user entity.User) error {
// 	return s.userRepository.Create(ctx, &user)
// }

func (s *wolService) UpdateUser(ctx context.Context, user entity.User) error {
	return s.userRepository.Update(ctx, &user)
}

func (s *wolService) DeleteUser(ctx context.Context, id int) error {
	return s.userRepository.Delete(ctx, id)
}
