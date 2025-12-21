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

func (r *UserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT COUNT(1) FROM users WHERE username = ? AND is_deleted = 0",
		username,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, role, is_deleted
		 FROM users WHERE username = ?`,
		username,
	)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.IsDeleted,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users
		(username, password_hash, role, is_deleted, created_at, updated_at)
		VALUES (?, ?, ?, 0, NOW(), NOW())`
	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.PasswordHash,
		user.Role,
	)
	return err
}
