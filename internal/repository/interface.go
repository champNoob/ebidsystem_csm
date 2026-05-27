package repository

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/repository/dto"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*model.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
}

type AdminRepository interface {
	GetGlobalStats(ctx context.Context) (
		totalOrders int64,
		totalTrades int64,
		totalVolume int64,
		totalTurnover float64,
		err error,
	)
	GetUserRanking(ctx context.Context, limit int) ([]dto.UserRank, error)
	GetUserRoleStats(ctx context.Context) ([]dto.UserRoleStat, error)
	GetSymbolStats(ctx context.Context) ([]dto.SymbolStat, error)
	GetOrderStatusStats(ctx context.Context) ([]dto.OrderStatusStat, error)
	GetRecentTrades(ctx context.Context, limit int) ([]dto.RecentTrade, error)
	GetTradeTimeline(ctx context.Context) ([]dto.TradeTimelinePoint, error)
	// 订单管理：
	GetAdminOrders(ctx context.Context, query dto.AdminOrderQuery) (*dto.AdminOrderPage, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) (uint64, error)
	FindByUserID(ctx context.Context, userID int64, statuses []model.OrderStatus) ([]*model.Order, error)
	FindAll(ctx context.Context, statuses []model.OrderStatus) ([]*model.Order, error)
	FindByID(ctx context.Context, id int64) (*model.Order, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	CancelOrder(ctx context.Context, orderID uint64) error
	//撮合事件事务化：
	WithTx(ctx context.Context, fn TxFunc) error
	FillOrderTx(ctx context.Context, tx *sql.Tx, symbol string, orderID uint64, qty int64) error
	CreateTradeTx(ctx context.Context, tx *sql.Tx, trade *model.Trade) error
	InsertMatchEventTx(ctx context.Context, tx *sql.Tx, eventID string, symbol string, buyOrderID uint64, sellOrderID uint64, quantity int64, price float64) (bool, error)
	// 重启恢复：
	FindActiveOrdersForRecovery(ctx context.Context) ([]*model.Order, error)
	FindDirtyOrdersForRecovery(ctx context.Context) ([]dto.RecoveryDirtyOrder, error)
}

type TxFunc func(tx *sql.Tx) error
