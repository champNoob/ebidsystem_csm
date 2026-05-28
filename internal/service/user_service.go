package service

import (
	"context"
	"ebidsystem_csm/internal/apperror"
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

	// 1. 密码长度校验：
	if len(input.Password) < 8 {
		return apperror.ErrPasswordTooShort
	} else if len(input.Password) > 50 {
		return apperror.ErrPasswordTooLong
	}
	// 2. 角色合法性校验：
	switch input.Role {
	case "client", "seller", "trader":
	default: //sales和admin不允许通过普通注册创建
		return apperror.ErrInvalidUserRole
	}

	// 3. 密码处理（业务规则）：
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return apperror.ErrInternal
	}
	//
	user := &model.User{
		Username:     input.Username,
		PasswordHash: hash,
		Role:         input.Role,
		IsDeleted:    false,
	}

	// 4. 调用仓储层：
	if err := s.repo.Create(ctx, user); err != nil {
		return apperror.ErrInternal
	}
	// 5. 创建审计日志
	// 6. 触发领域事件

	return nil
}

type LoginInput struct {
	Username string
	Password string
}

func (s *UserService) Login(ctx context.Context, input LoginInput) (string, error) {
	user, err := s.repo.FindByUsername(ctx, input.Username)
	if err != nil {
		return "", apperror.ErrInternal
	}
	if user == nil || user.IsDeleted {
		return "", apperror.ErrUserNotFound
	}

	if !security.VerifyPassword(input.Password, user.PasswordHash) {
		return "", apperror.ErrInvalidPassword
	}

	// 生成 JWT（下一步）
	token, err := security.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return "", apperror.ErrInternal
	}

	return token, nil
}
