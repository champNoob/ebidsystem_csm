package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/service"
	"fmt"
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
			&o.FilledQuantity,
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
		`SELECT id, user_id, symbol, side, price, quantity, filled_quantity, status, created_at
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

func (r *OrderRepo) FillOrder(
	ctx context.Context,
	orderID uint64,
	filledQty int64,
) error {
	log.Printf( //x
		"[FILL_ORDER] orderID=%d filledQty=%d",
		orderID, filledQty,
	)
	// 先查询订单当前状态和数量，用于更精确的错误信息
	var currentStatus string
	var filledQuantity, quantity int64
	err := r.db.QueryRowContext(ctx,
		"SELECT status, filled_quantity, quantity FROM orders WHERE id = ?",
		orderID,
	).Scan(&currentStatus, &filledQuantity, &quantity)
	if err == sql.ErrNoRows {
		return service.ErrOrderNotFound
	}
	if err != nil {
		return err
	}
	// 添加详细日志
	log.Printf(
		"[FILL_ORDER_DETAIL] orderID=%d, currentStatus=%s, filledQuantity=%d, quantity=%d, filledQty=%d, newTotal=%d",
		orderID, currentStatus, filledQuantity, quantity, filledQty, filledQuantity+filledQty,
	)
	// 检查订单状态
	if currentStatus == "cancelled" || currentStatus == "filled" {
		return fmt.Errorf("order %d is already %s", orderID, currentStatus)
	}
	// 检查是否会超额填充
	if filledQuantity+filledQty > quantity {
		return service.ErrOrderOverFilled
	}

	newFilled := filledQuantity + filledQty
	var newStatus string
	if newFilled >= quantity {
		newStatus = string(model.OrderStatusFilled)
	} else {
		newStatus = string(model.OrderStatusPartial)
	}
	res, err := r.db.ExecContext(
		ctx,
		`UPDATE orders
		SET filled_quantity = ?, status = ?, updated_at = NOW()
		WHERE id = ? AND status IN ('pending', 'partial')`,
		newFilled,
		newStatus,
		orderID,
	)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("[FILL_ORDER_ERROR] RowsAffected error: %v", err)
		return err
	}
	log.Printf("[FILL_ORDER_RESULT] RowsAffected: %d", rows)
	if rows == 0 {
		// 再查询一次看看状态是否变了
		var afterStatus string
		var afterFilledQuantity int64
		r.db.QueryRowContext(ctx,
			"SELECT status, filled_quantity FROM orders WHERE id = ?",
			orderID,
		).Scan(&afterStatus, &afterFilledQuantity)
		log.Printf("[FILL_ORDER_AFTER] orderID=%d, status=%s, filled_quantity=%d",
			orderID, afterStatus, afterFilledQuantity)
		return service.ErrOrderUpdateFailed
	}
	// 查询更新后的状态
	var finalStatus string
	var finalFilledQuantity int64
	r.db.QueryRowContext(ctx,
		"SELECT status, filled_quantity FROM orders WHERE id = ?",
		orderID,
	).Scan(&finalStatus, &finalFilledQuantity)
	log.Printf("[FILL_ORDER_FINAL] orderID=%d, status=%s, filled_quantity=%d",
		orderID, finalStatus, finalFilledQuantity)
	return nil
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

func (r *OrderRepo) CancelOrder(
	ctx context.Context,
	orderID uint64,
) error {
	res, err := r.db.ExecContext(
		ctx,
		`
		UPDATE orders
		SET
			status = 'cancelled',
			updated_at = NOW()
		WHERE id = ?
		  AND status IN ('pending', 'partial');
		`,
		orderID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return service.ErrOrderNotCancellable
	}

	return nil
}
