package mysql

import (
	"context"
	"database/sql"
)

func (r *OrderRepo) TryInsertMatchEventTx(
	ctx context.Context,
	tx *sql.Tx,
	eventID string,
) (bool, error) {

	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO match_events (event_id) VALUES (?)`,
		eventID,
	)

	if err != nil {
		if isMySQLDuplicateEntry(err) {
			return false, nil // 已处理过
		}
		return false, err
	}

	return true, nil
}
