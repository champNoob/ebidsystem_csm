package mysql

import (
	"context"
	"database/sql"
)

func (r *OrderRepo) InsertMatchEventTx( //事件去重
	ctx context.Context,
	tx *sql.Tx,
	eventID string,
	symbol string,
	buyOrderID uint64,
	sellOrderID uint64,
	quantity int64,
	price float64,
) (bool, error) {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO match_events (event_id, symbol, buy_order_id, sell_order_id, quantity, price) VALUES (?, ?, ?, ?, ?, ?)`,
		eventID,
		symbol,
		buyOrderID,
		sellOrderID,
		quantity,
		price,
	)
	if err != nil {
		if isMySQLDuplicateEntry(err) {
			return false, nil // 已处理过
		}
		return false, err
	}
	return true, nil // 首次处理
}
