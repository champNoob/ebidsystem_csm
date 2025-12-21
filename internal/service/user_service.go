package service

import (
	"context"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/pkg/security"
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

type CreateUserInput struct {
	Username string
	Password string
	Role     string
}

func (s *UserService) CreateUser(
	ctx context.Context,
	input CreateUserInput,
) error {

	// 1. 密码处理（业务规则）
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     input.Username,
		PasswordHash: hash,
		Role:         input.Role,
		IsDeleted:    false,
	}

	// 2. username 唯一性校验
	// 3. role 合法性校验
	// 4. 创建审计日志
	// 5. 触发领域事件

	return s.repo.Create(ctx, user)
}
