package repository

import (
	"context"
	"ebidsystem_csm/internal/model"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) (uint64, error)
	FindByUserID(ctx context.Context, userID int64) ([]*model.Order, error)
	FindAll(ctx context.Context) ([]*model.Order, error)
	FindByID(ctx context.Context, id int64) (*model.Order, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	FillOrder(ctx context.Context, orderID uint64, filledQty int64) error
	CreateTrade(ctx context.Context, trade *model.Trade) error
}
