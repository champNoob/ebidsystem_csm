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
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
