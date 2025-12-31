# MySQL 的恶性 BUG

出现问题的仓储层代码：

```go
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

	res, err := r.db.ExecContext(
		ctx,
		`
		UPDATE orders
		SET
			filled_quantity = filled_quantity + ?,
			status = CASE
				WHEN (filled_quantity + ?) >= quantity THEN 'filled'
				ELSE 'partial'
			END,
			updated_at = NOW()
		WHERE id = ?
			AND status IN ('pending', 'partial')
			AND (filled_quantity + ?) <= quantity;
		`,
		filledQty,
		filledQty,
		orderID,
		filledQty,
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
```

执行时没有任何报错，但执行后：


卖方订单的quantity为10，filled_quantity为7，但它的status却被置为了filled，应为partitial。
买方订单quantity为7，filled_quantity为7，status为filled，这个是正常的。

控制台日志：

```txt
2025/12/30 19:55:05 [FILL_ORDER] orderID=75 filledQty=7
2025/12/30 19:55:05 [FILL_ORDER_DETAIL] orderID=75, currentStatus=pending, filledQuantity=0, quantity=7, filledQty=7, newTotal=7
2025/12/30 19:55:05 [FILL_ORDER_RESULT] RowsAffected: 1
2025/12/30 19:55:05 [FILL_ORDER_FINAL] orderID=75, status=filled, filled_quantity=7
2025/12/30 19:55:05 [FILL_ORDER] orderID=74 filledQty=7
2025/12/30 19:55:05 [FILL_ORDER_DETAIL] orderID=74, currentStatus=pending, filledQuantity=0, quantity=10, filledQty=7, newTotal=7
2025/12/30 19:55:05 [FILL_ORDER_RESULT] RowsAffected: 1
2025/12/30 19:55:05 [FILL_ORDER_FINAL] orderID=74, status=filled, filled_quantity=7
```

但将 MySQL 语句部分改为：

```go
res, err := r.db.ExecContext(
	ctx,
	`UPDATE orders
	SET filled_quantity = ?, status = ?, updated_at = NOW()
	WHERE id = ? AND status IN ('pending', 'partial')`,
	newFilled,
	newStatus,
	orderID,
)
```

就成功了……
