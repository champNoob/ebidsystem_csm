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

func (r *OrderRepo) Create(ctx context.Context, o *model.Order) (uint64, error) {
	query := `
INSERT INTO orders (user_id, symbol, side, price, quantity, status)
VALUES (?, ?, ?, ?, ?, ?)
`
	result, err := r.db.ExecContext(
		ctx,
		query,
		o.UserID,
		o.Symbol,
		o.Side,
		o.Price,
		o.Quantity,
		o.Status,
	)
	if err != nil {
		return 0, err
	}
	// 获取新插入的 ID：
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
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
			&o.UserID,
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
			&o.UserID,
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

func (r *OrderRepo) FindByID(ctx context.Context, id int64) (*model.Order, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, symbol, side, price, quantity, status, created_at
		 FROM orders WHERE id = ?`,
		id,
	)

	var o model.Order
	if err := row.Scan(
		&o.ID,
		&o.UserID,
		&o.Symbol,
		&o.Side,
		&o.Price,
		&o.Quantity,
		&o.Status,
		&o.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &o, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE orders SET status = ? WHERE id = ?`,
		status,
		id,
	)
	return err
}
