package mysql

import (
	"context"
	"database/sql"
)

func (r *OrderRepo) WithTx(
	ctx context.Context,
	fn func(tx *sql.Tx) error,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
