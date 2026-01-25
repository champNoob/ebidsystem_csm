package service

import (
	"context"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/pkg/security"
	"ebidsystem_csm/internal/repository"
	"strings"
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
		return ErrPasswordTooShort
	}
	// 2. 角色合法性校验：
	switch input.Role {
	case "client", "seller", "trader", "admin":
	default:
		return ErrInvalidUserRole
	}

	// 3. 密码处理（业务规则）：
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return err
	}
	//
	user := &model.User{
		Username:     input.Username,
		PasswordHash: hash,
		Role:         input.Role,
		IsDeleted:    false,
	}

	// 4. 用户名唯一性校验：
	if err := s.repo.Create(ctx, user); err != nil {
		// MySQL 错误 1062 → 唯一键冲突
		if isMySQLDuplicateEntry(err) {
			return ErrUserAlreadyExists
		}
		return ErrInternal
	}
	// 5. 创建审计日志
	// 6. 触发领域事件

	return s.repo.Create(ctx, user)
}

func isMySQLDuplicateEntry(err error) bool {
	// 简单匹配 MySQL 错误号 1062
	return strings.Contains(err.Error(), "Error 1062")
}

type LoginInput struct {
	Username string
	Password string
}

func (s *UserService) Login(ctx context.Context, input LoginInput) (string, error) {
	user, err := s.repo.FindByUsername(ctx, input.Username)
	if err != nil {
		return "", ErrInternal
	}
	if user == nil || user.IsDeleted {
		return "", ErrUserNotFound
	}

	if !security.VerifyPassword(input.Password, user.PasswordHash) {
		return "", ErrInvalidPassword
	}

	// 生成 JWT（下一步）
	token, err := security.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return "", ErrInternal
	}

	return token, nil
}
