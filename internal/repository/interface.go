package repository

import (
	"context"
	"ebidsystem_csm/internal/model"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
}
