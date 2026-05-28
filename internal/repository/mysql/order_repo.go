package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/apperror"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/repository"
	"ebidsystem_csm/internal/repository/dto"
	"log"
	"strings"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) repository.OrderRepository {
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
		return 0, wrapDBError(err)
	}
	// 获取新插入的 ID：
	id, err := result.LastInsertId()
	if err != nil {
		return 0, wrapDBError(err)
	}
	return uint64(id), nil
}

// 动态绑定 IN (...)：
func buildStatusCondition(statuses []model.OrderStatus) (string, []interface{}) {
	if len(statuses) == 0 {
		return "", nil
	}

	placeholders := make([]string, len(statuses))
	args := make([]interface{}, len(statuses))

	for i, s := range statuses {
		placeholders[i] = "?"
		args[i] = s
	}

	return " AND status IN (" + strings.Join(placeholders, ",") + ")", args
}

func (r *OrderRepo) FindByUserID(
	ctx context.Context,
	userID int64,
	statuses []model.OrderStatus,
) ([]*model.Order, error) {
	query := `
	SELECT id, user_id, symbol, side, price, quantity, filled_quantity, status, created_at
	FROM orders
	WHERE user_id = ?
	`

	args := []interface{}{userID}

	cond, condArgs := buildStatusCondition(statuses)
	query += cond
	args = append(args, condArgs...)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, wrapDBError(err)
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
			return nil, wrapDBError(err)
		}
		orders = append(orders, &o)
	}

	return orders, nil
}

func (r *OrderRepo) FindAll(
	ctx context.Context,
	statuses []model.OrderStatus,
) ([]*model.Order, error) {
	query := `
	SELECT id, user_id, symbol, side, price, quantity, filled_quantity, status, created_at
	FROM orders
	WHERE 1=1
	`
	args := []interface{}{}

	cond, condArgs := buildStatusCondition(statuses)
	query += cond
	args = append(args, condArgs...)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		// log.Printf("%v", err)//#
		return nil, wrapDBError(err)
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
			return nil, wrapDBError(err)
		}
		orders = append(orders, &o)
	}

	return orders, nil
}

func (r *OrderRepo) FindByID(
	ctx context.Context,
	id int64,
) (*model.Order, error) {
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
		if err == sql.ErrNoRows {
			return nil, apperror.ErrOrderNotFound
		}
		return nil, wrapDBError(err)
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
	return wrapDBError(err)
}

func (r *OrderRepo) FillOrderTx(
	ctx context.Context,
	tx *sql.Tx,
	symbol string,
	orderID uint64,
	matchQty int64,
) error {
	var currentSymbol string
	var status string
	var filledQuantity int64
	var quantity int64

	err := tx.QueryRowContext(
		ctx,
		`
		SELECT symbol, status, filled_quantity, quantity
		FROM orders
		WHERE id = ?
		FOR UPDATE
		`,
		orderID,
	).Scan(&currentSymbol, &status, &filledQuantity, &quantity)

	if err == sql.ErrNoRows {
		log.Printf(
			"[FILL_ORDER_TX_NOT_FOUND] orderID=%d symbol=%s matchQty=%d",
			orderID,
			symbol,
			matchQty,
		)
		return apperror.ErrOrderNotFound
	}
	if err != nil {
		log.Printf(
			"[FILL_ORDER_TX_SELECT_ERROR] orderID=%d symbol=%s err=%v",
			orderID,
			symbol,
			err,
		)
		return wrapDBError(err)
	}

	if currentSymbol != symbol {
		log.Printf(
			"[FILL_ORDER_TX_SYMBOL_MISMATCH] orderID=%d dbSymbol=%s eventSymbol=%s status=%s filled=%d quantity=%d matchQty=%d",
			orderID,
			currentSymbol,
			symbol,
			status,
			filledQuantity,
			quantity,
			matchQty,
		)
		return apperror.ErrOrderSymbolMismatch
	}

	if status != string(model.OrderStatusPending) &&
		status != string(model.OrderStatusPartial) {
		log.Printf(
			"[FILL_ORDER_TX_STATUS_NOT_FILLABLE] orderID=%d symbol=%s status=%s filled=%d quantity=%d matchQty=%d",
			orderID,
			symbol,
			status,
			filledQuantity,
			quantity,
			matchQty,
		)
		return apperror.ErrOrderNotFillable
	}

	if matchQty <= 0 {
		log.Printf(
			"[FILL_ORDER_TX_INVALID_QTY] orderID=%d symbol=%s status=%s filled=%d quantity=%d matchQty=%d",
			orderID,
			symbol,
			status,
			filledQuantity,
			quantity,
			matchQty,
		)
		return apperror.ErrMatchEventInvalid
	}

	newFilled := filledQuantity + matchQty
	if newFilled > quantity {
		log.Printf(
			"[FILL_ORDER_TX_OVER_FILLED] orderID=%d symbol=%s status=%s filled=%d quantity=%d matchQty=%d newFilled=%d",
			orderID,
			symbol,
			status,
			filledQuantity,
			quantity,
			matchQty,
			newFilled,
		)
		return apperror.ErrOrderOverFilled
	}

	newStatus := string(model.OrderStatusPartial)
	if newFilled == quantity {
		newStatus = string(model.OrderStatusFilled)
	}

	res, err := tx.ExecContext(
		ctx,
		`
		UPDATE orders
		SET
			filled_quantity = ?,
			status = ?,
			updated_at = NOW()
		WHERE id = ?
		`,
		newFilled,
		newStatus,
		orderID,
	)
	if err != nil {
		log.Printf(
			"[FILL_ORDER_TX_UPDATE_ERROR] orderID=%d symbol=%s oldFilled=%d newFilled=%d quantity=%d oldStatus=%s newStatus=%s err=%v",
			orderID,
			symbol,
			filledQuantity,
			newFilled,
			quantity,
			status,
			newStatus,
			err,
		)
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf(
			"[FILL_ORDER_TX_ROWS_AFFECTED_ERROR] orderID=%d symbol=%s err=%v",
			orderID,
			symbol,
			err,
		)
		return err
	}

	if rows == 0 {
		log.Printf(
			"[FILL_ORDER_TX_NO_ROWS_AFFECTED] orderID=%d symbol=%s oldFilled=%d newFilled=%d quantity=%d oldStatus=%s newStatus=%s",
			orderID,
			symbol,
			filledQuantity,
			newFilled,
			quantity,
			status,
			newStatus,
		)
		return apperror.ErrOrderUpdateFailed
	}

	return nil
}

func (r *OrderRepo) CreateTradeTx(
	ctx context.Context,
	tx *sql.Tx,
	trade *model.Trade,
) error {
	_, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO trades (event_id, symbol, buy_order_id, sell_order_id, price, quantity)
		VALUES (?, ?, ?, ?, ?, ?)
		`,
		trade.EventID,
		trade.Symbol,
		trade.BuyOrderID,
		trade.SellOrderID,
		trade.Price,
		trade.Quantity,
	)
	if err != nil {
		if isMySQLDuplicateEntry(err) { //幂等命中，该撮合事件已处理
			return nil
		}
		// log.Println("order_repo CreateTradeTx error:", err)//
		return wrapDBError(err)
	}
	return nil
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
		return wrapDBError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return wrapDBError(err)
	}

	if rows == 0 {
		return apperror.ErrOrderNotCancellable
	}

	return nil
}

/* 引擎重启恢复订单 */

func (r *OrderRepo) FindActiveOrdersForRecovery(ctx context.Context) ([]*model.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, symbol, type, side, price, quantity, filled_quantity, status, created_at
		FROM orders
		WHERE status IN ('pending', 'partial')
			AND filled_quantity < quantity
			AND type = 'limit'
			AND price IS NOT NULL
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		var price float64
		if err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.Symbol,
			&o.Type,
			&o.Side,
			&price,
			&o.Quantity,
			&o.FilledQuantity,
			&o.Status,
			&o.CreatedAt,
		); err != nil {
			return nil, wrapDBError(err)
		}
		o.Price = &price
		orders = append(orders, &o)
	}

	if err = rows.Err(); err != nil {
		return nil, wrapDBError(err)
	}

	return orders, nil
}

func (r *OrderRepo) FindDirtyOrdersForRecovery(
	ctx context.Context,
) ([]dto.RecoveryDirtyOrder, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			symbol,
			status,
			quantity,
			filled_quantity,
			CASE
				WHEN filled_quantity > quantity THEN 'filled_quantity_gt_quantity'
				WHEN status = 'filled' AND filled_quantity < quantity THEN 'filled_status_but_not_full'
				WHEN status IN ('pending', 'partial') AND filled_quantity >= quantity THEN 'active_status_but_already_full'
				WHEN filled_quantity < 0 THEN 'negative_filled_quantity'
				ELSE 'unknown'
			END AS reason
		FROM orders
		WHERE
			filled_quantity > quantity
			OR (status = 'filled' AND filled_quantity < quantity)
			OR (status IN ('pending', 'partial') AND filled_quantity >= quantity)
			OR filled_quantity < 0
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, wrapDBError(err)
	}
	defer rows.Close()

	res := make([]dto.RecoveryDirtyOrder, 0)

	for rows.Next() {
		var item dto.RecoveryDirtyOrder
		if err := rows.Scan(
			&item.ID,
			&item.Symbol,
			&item.Status,
			&item.Quantity,
			&item.FilledQuantity,
			&item.Reason,
		); err != nil {
			return nil, wrapDBError(err)
		}

		res = append(res, item)
	}

	if err := rows.Err(); err != nil {
		return nil, wrapDBError(err)
	}

	return res, nil
}
