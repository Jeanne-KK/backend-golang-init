package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(
	users repository.UserRepository,
) *UserService {
	return &UserService{
		users: users,
	}
}

func (s *UserService) FindAll(ctx context.Context) (*[]model.User, error) {
	return s.users.FindAll(ctx)
}
