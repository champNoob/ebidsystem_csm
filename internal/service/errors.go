package service

import "errors"

var (
	// 用户侧：
	ErrUserAlreadyExists = errors.New("该用户名已经存在")   //“user already exists”
	ErrInvalidInput      = errors.New("输入错误")       //“invalid input”
	ErrInternal          = errors.New("内部错误，请稍后重试") //“internal error”
	ErrUserNotFound      = errors.New("找不到该用户")     //“user not found”
	ErrRoleInvalid       = errors.New("角色不合法")      //“invalid role”
	ErrInvalidPassword   = errors.New("密码错误")       //“invalid password”
	ErrPasswordTooShort  = errors.New("密码长度不足 8 位") //“the length of the password is too short”
	// 订单侧：
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderNotCancellable = errors.New("order is not cancellable")
	ErrOrderOverFilled     = errors.New("order is over filled")
	ErrOrderUpdateFailed   = errors.New("order update failed")
	ErrPermissionDenied    = errors.New("permission denied")
)
