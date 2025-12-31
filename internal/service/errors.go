package service

import "errors"

var (
	// 用户侧：
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternal          = errors.New("internal error")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	// 订单侧：
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderNotCancellable = errors.New("order is not cancellable")
	ErrOrderOverFilled     = errors.New("order is over filled")
	ErrOrderUpdateFailed   = errors.New("order update failed")
	ErrPermissionDenied    = errors.New("permission denied")
)
