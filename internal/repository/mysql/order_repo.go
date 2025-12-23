package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/model"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, o *model.Order) error {
	query := `
INSERT INTO orders (user_id, symbol, side, price, quantity, status)
VALUES (?, ?, ?, ?, ?, ?)
`
	_, err := r.db.ExecContext(
		ctx,
		query,
		o.CreatorID,
		o.Symbol,
		o.Side,
		o.Price,
		o.Quantity,
		o.Status,
	)
	return err
}

func (r *OrderRepo) FindByUserID(ctx context.Context, userID int64) ([]*model.Order, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, user_id, symbol, side, price, quantity, status, created_at
		 FROM orders WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.ID,
			&o.CreatorID,
			&o.Symbol,
			&o.Side,
			&o.Price,
			&o.Quantity,
			&o.Status,
			&o.CreatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}

func (r *OrderRepo) FindAll(ctx context.Context) ([]*model.Order, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, user_id, symbol, side, price, quantity, status, created_at FROM orders`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.ID,
			&o.CreatorID,
			&o.Symbol,
			&o.Side,
			&o.Price,
			&o.Quantity,
			&o.Status,
			&o.CreatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, nil
}
