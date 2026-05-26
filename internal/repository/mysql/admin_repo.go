package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/repository"
	"ebidsystem_csm/internal/repository/dto"
)

type AdminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) repository.AdminRepository {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) GetGlobalStats(ctx context.Context) (
	totalOrders int64,
	totalTrades int64,
	totalVolume int64,
	totalTurnover float64,
	err error,
) {
	// 订单数
	err = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM orders`,
	).Scan(&totalOrders)
	if err != nil {
		return
	}

	// 成交统计
	err = r.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*),
			IFNULL(SUM(quantity),0),
			IFNULL(SUM(price * quantity),0)
		FROM trades
	`).Scan(&totalTrades, &totalVolume, &totalTurnover)

	return
}

func (r *AdminRepo) GetUserRoleStats(ctx context.Context) ([]dto.UserRoleStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT role, COUNT(*)
		FROM users
		WHERE is_deleted = 0
		GROUP BY role
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []dto.UserRoleStat
	for rows.Next() {
		var s dto.UserRoleStat
		if err := rows.Scan(&s.Role, &s.Count); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}

func (r *AdminRepo) GetUserRanking(ctx context.Context, limit int) ([]dto.UserRank, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			u.id,
			u.username,
			u.role,
			IFNULL(SUM(x.buy_volume), 0) AS buy_volume,
			IFNULL(SUM(x.sell_volume), 0) AS sell_volume,
			IFNULL(SUM(x.buy_volume + x.sell_volume), 0) AS total_volume
		FROM users u
		JOIN (
			SELECT 
				ob.user_id AS user_id,
				SUM(t.quantity) AS buy_volume,
				0 AS sell_volume
			FROM trades t
			JOIN orders ob ON t.buy_order_id = ob.id
			GROUP BY ob.user_id

			UNION ALL

			SELECT 
				os.user_id AS user_id,
				0 AS buy_volume,
				SUM(t.quantity) AS sell_volume
			FROM trades t
			JOIN orders os ON t.sell_order_id = os.id
			GROUP BY os.user_id
		) x ON u.id = x.user_id
		WHERE u.is_deleted = 0
		GROUP BY u.id, u.username, u.role
		ORDER BY total_volume DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []dto.UserRank
	for rows.Next() {
		var u dto.UserRank
		if err := rows.Scan(
			&u.UserID,
			&u.Username,
			&u.Role,
			&u.BuyVolume,
			&u.SellVolume,
			&u.TotalVolume,
		); err != nil {
			return nil, err
		}
		res = append(res, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *AdminRepo) GetSymbolStats(ctx context.Context) ([]dto.SymbolStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			symbol,
			COUNT(*) as trades,
			SUM(quantity) as volume,
			SUM(price * quantity) as turnover
		FROM trades
		GROUP BY symbol
		ORDER BY volume DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []dto.SymbolStat
	for rows.Next() {
		var s dto.SymbolStat
		if err := rows.Scan(
			&s.Symbol,
			&s.Trades,
			&s.Volume,
			&s.Turnover,
		); err != nil {
			return nil, err
		}
		res = append(res, s)
	}

	return res, nil
}

func (r *AdminRepo) GetOrderStatusStats(ctx context.Context) ([]dto.OrderStatusStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM orders
		GROUP BY status
		ORDER BY COUNT(*) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []dto.OrderStatusStat
	for rows.Next() {
		var s dto.OrderStatusStat
		if err := rows.Scan(&s.Status, &s.Count); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}

func (r *AdminRepo) GetRecentTrades(ctx context.Context, limit int) ([]dto.RecentTrade, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_id, buy_order_id, sell_order_id, price, quantity, created_at
		FROM trades
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []dto.RecentTrade
	for rows.Next() {
		var t dto.RecentTrade
		if err := rows.Scan(
			&t.ID,
			&t.EventID,
			&t.BuyOrderID,
			&t.SellOrderID,
			&t.Price,
			&t.Quantity,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (r *AdminRepo) GetTradeTimeline(ctx context.Context) ([]dto.TradeTimelinePoint, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			DATE_FORMAT(created_at, '%Y-%m-%d %H:%i') AS time_bucket,
			COUNT(*) AS trades,
			IFNULL(SUM(quantity), 0) AS volume,
			IFNULL(SUM(price * quantity), 0) AS turnover
		FROM trades
		GROUP BY time_bucket
		ORDER BY time_bucket ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []dto.TradeTimelinePoint
	for rows.Next() {
		var p dto.TradeTimelinePoint
		if err := rows.Scan(&p.TimeBucket, &p.Trades, &p.Volume, &p.Turnover); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}
