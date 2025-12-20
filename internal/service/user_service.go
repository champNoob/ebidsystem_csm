package service

import (
	"context"
	"ebidsystem_csm/internal/api/dto/request"
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

func (s *UserService) CreateUser(
	ctx context.Context,
	req request.CreateUserRequest,
) error {

	// 1. 密码处理（业务规则）
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Role:         req.Role,
		IsDeleted:    false,
	}

	// 2. username 唯一性校验
	// 3. role 合法性校验
	// 4. 创建审计日志
	// 5. 触发领域事件

	return s.repo.Create(ctx, user)
}
