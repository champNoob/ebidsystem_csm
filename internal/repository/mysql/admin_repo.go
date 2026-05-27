package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/repository"
	"ebidsystem_csm/internal/repository/dto"
	"fmt"
	"strings"
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
		LEFT JOIN (
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
			IFNULL(symbol, 'UNKNOWN') AS symbol,
			COUNT(*) as trades,
			FNULL(SUM(quantity), 0) as volume,
			IFNULL(SUM(price * quantity), 0) as turnover
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
		SELECT
			id,
			event_id,
			IFNULL(symbol, ''),
			buy_order_id,
			sell_order_id,
			price,
			quantity,
			DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')
		FROM trades
		ORDER BY created_at DESC, id DESC
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
			&t.Symbol,
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

func (r *AdminRepo) GetAdminOrders(
	ctx context.Context,
	query dto.AdminOrderQuery,
) (*dto.AdminOrderPage, error) {
	whereParts := []string{"1=1"}
	args := make([]interface{}, 0)

	if query.UserID != nil {
		whereParts = append(whereParts, "user_id = ?")
		args = append(args, *query.UserID)
	}

	if query.Symbol != "" {
		whereParts = append(whereParts, "symbol = ?")
		args = append(args, query.Symbol)
	}

	if query.Status != "" && query.Status != "all" {
		whereParts = append(whereParts, "status = ?")
		args = append(args, query.Status)
	}

	if query.Side != "" && query.Side != "all" {
		whereParts = append(whereParts, "side = ?")
		args = append(args, query.Side)
	}

	if query.Type != "" && query.Type != "all" {
		whereParts = append(whereParts, "`type` = ?")
		args = append(args, query.Type)
	}

	whereSQL := strings.Join(whereParts, " AND ")

	var total int64
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM orders
		WHERE %s
	`, whereSQL)

	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (query.Page - 1) * query.PageSize

	listArgs := append([]interface{}{}, args...)
	listArgs = append(listArgs, query.PageSize, offset)

	listSQL := fmt.Sprintf(`
		SELECT
			id,
			user_id,
			symbol,
			`+"`type`"+`,
			side,
			price,
			quantity,
			filled_quantity,
			status,
			DATE_FORMAT(created_at, '%%Y-%%m-%%d %%H:%%i:%%s')
		FROM orders
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	rows, err := r.db.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dto.AdminOrderItem, 0)

	for rows.Next() {
		var item dto.AdminOrderItem
		var price sql.NullFloat64

		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Symbol,
			&item.Type,
			&item.Side,
			&price,
			&item.Quantity,
			&item.FilledQuantity,
			&item.Status,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}

		if price.Valid {
			item.Price = &price.Float64
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &dto.AdminOrderPage{
		Items:    items,
		Page:     query.Page,
		PageSize: query.PageSize,
		Total:    total,
	}, nil
}
