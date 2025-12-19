package service

import (
	"context"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}
