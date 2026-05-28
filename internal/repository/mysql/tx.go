package mysql

import (
	"context"
	"ebidsystem_csm/internal/repository"
)

func (r *OrderRepo) WithTx(
	ctx context.Context,
	fn repository.TxFunc,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return wrapDBError(err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return wrapDBError(err)
	}

	if err := tx.Commit(); err != nil {
		return wrapDBError(err)
	}

	return nil
}
