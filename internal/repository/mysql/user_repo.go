package mysql

import (
	"context"
	"database/sql"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/repository"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) repository.UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, username FROM users WHERE id = ?",
		id,
	)

	var u model.User
	if err := row.Scan(&u.ID, &u.Username); err != nil {
		return nil, err
	}
	return &u, nil
}
