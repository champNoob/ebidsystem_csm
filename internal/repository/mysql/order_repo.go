package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/model"
	"log"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, o *model.Order) (uint64, error) {
	query := `
INSERT INTO orders (user_id, symbol, side, price, quantity, filled_quantity, status)
VALUES (?, ?, ?, ?, ?, 0, ?)
`
	result, err := r.db.ExecContext(
		ctx,
		query,
		o.UserID,
		o.Symbol,
		o.Side,
		o.Price,
		o.Quantity,
		model.OrderStatusPending, // 使用定义好的常量
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
		`SELECT id, user_id, symbol, side, price, quantity, filled_quantity, status, created_at
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
		&o.FilledQuantity,
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

func (r *OrderRepo) FillOrder(ctx context.Context, orderID uint64, filledQty int64) error {
	_, err := r.db.ExecContext(
		ctx,
		// "filled_quantity + ?" 必须写两次（MySQL 不支持引用更新后的列）：
		`UPDATE orders
		SET
		  filled_quantity = filled_quantity + ?,
		  status = CASE
		    WHEN filled_quantity + ? >= quantity THEN 'filled'
		    ELSE 'partial'
		  END
		WHERE id = ? AND status IN ('pending', 'partial');`,
		filledQty,
		filledQty,
		orderID,
	)
	if err != nil {
		log.Printf("[DB_ERROR] update order %d failed: %v", orderID, err)
		return err
	}
	return err
}

func (r *OrderRepo) CreateTrade(ctx context.Context, trade *model.Trade) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO trades (buy_order_id, sell_order_id, price, quantity) VALUES (?, ?, ?, ?)`,
		trade.BuyOrderID,
		trade.SellOrderID,
		trade.Price,
		trade.Quantity,
	)
	return err
}
