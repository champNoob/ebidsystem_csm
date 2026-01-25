package service

import "errors"

var (
	ErrInvalidInput     = errors.New("输入错误")       //“invalid input”
	ErrInternal         = errors.New("内部错误，请稍后重试") //“internal error”
	ErrPermissionDenied = errors.New("权限不足")       //“permission denied”
	// 用户侧：
	ErrUserAlreadyExists = errors.New("该用户名已经存在")   //“user already exists”
	ErrUserNotFound      = errors.New("找不到该用户")     //“user not found”
	ErrRoleInvalid       = errors.New("角色不合法")      //“invalid role”
	ErrInvalidPassword   = errors.New("密码错误")       //“invalid password”
	ErrPasswordTooShort  = errors.New("密码长度不足 8 位") //“the length of the password is too short”
	ErrRoleSideMismatch  = errors.New("角色与订单方向不匹配") //“role side mismatch”
	// 订单侧：
	ErrOrderNotFound           = errors.New("找不到该订单")    //“order not found”
	ErrOrderNotCancellable     = errors.New("订单无法取消")    //“order is not cancellable”
	ErrOrderOverFilled         = errors.New("订单已超额")     //“order is over filled”
	ErrOrderUpdateFailed       = errors.New("订单更新失败")    //“order update failed”
	ErrOrderLimitWithoutPrice  = errors.New("限价单需要价格")   //“limit order requires price”
	ErrOrderMarketWithPrice    = errors.New("市价单不能有价格")  //“market order must not have price”
	ErrOrderInvalidStatusQuery = errors.New("无效的订单状态查询") //“invalid order status query”
	ErrOrderInvalidType        = errors.New("无效的订单类型")   //“invalid order type”
)
