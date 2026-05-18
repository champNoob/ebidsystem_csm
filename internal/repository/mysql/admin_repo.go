package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/repository/dto"
)

type AdminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) *AdminRepo {
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

func (r *AdminRepo) GetSymbolStats(ctx context.Context) ([]dto.SymbolStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			symbol,
			COUNT(*) as trades,
			SUM(quantity) as volume,
			SUM(price * quantity) as turnover
		FROM trades
		GROUP BY symbol
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

func (r *AdminRepo) GetUserRanking(ctx context.Context, limit int) ([]dto.UserRank, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.user_id, SUM(t.quantity) as volume
		FROM trades t
		JOIN orders o ON t.buy_order_id = o.id
		GROUP BY o.user_id
		ORDER BY volume DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []dto.UserRank
	for rows.Next() {
		var u dto.UserRank
		if err := rows.Scan(&u.UserID, &u.Volume); err != nil {
			return nil, err
		}
		res = append(res, u)
	}

	return res, nil
}
