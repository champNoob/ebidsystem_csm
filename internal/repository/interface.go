package repository

import (
	"context"
	"database/sql"
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
	FindByUserID(ctx context.Context, userID int64, statuses []model.OrderStatus) ([]*model.Order, error)
	FindAll(ctx context.Context, statuses []model.OrderStatus) ([]*model.Order, error)
	FindByID(ctx context.Context, id int64) (*model.Order, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	FillOrder(ctx context.Context, orderID uint64, filledQty int64) error //不用于撮合事件！
	CancelOrder(ctx context.Context, orderID uint64) error
	CreateTrade(ctx context.Context, trade *model.Trade) error //不用于撮合事件！
	//撮合事件事务化
	WithTx(ctx context.Context, fn TxFunc) error
	FillOrderTx(ctx context.Context, tx *sql.Tx, orderID uint64, qty int64) error
	CreateTradeTx(ctx context.Context, tx *sql.Tx, trade *model.Trade) error
}

type TxFunc func(tx *sql.Tx) error
