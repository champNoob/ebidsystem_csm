package service

// BusinessError 业务错误结构体
type BusinessError struct {
	Code    string
	Message string
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	return e.Message
}

var (
	// 通用：
	ErrInvalidInput = &BusinessError{
		Code:    "INVALID_INPUT",
		Message: "输入错误",
	}
	ErrInternal = &BusinessError{
		Code:    "INTERNAL_ERROR",
		Message: "内部错误，请稍后重试",
	}
	//登录与鉴权：
	ErrMissingAuthHeader = &BusinessError{
		Code:    "AUTH_MISSING_HEADER",
		Message: "缺少认证头",
	}
	ErrInvalidAuthHeader = &BusinessError{
		Code:    "AUTH_INVALID_HEADER",
		Message: "认证头不合法",
	}
	ErrInvalidToken = &BusinessError{
		Code:    "AUTH_INVALID_TOKEN",
		Message: "token不合法",
	}
	ErrInvalidTokenClaims = &BusinessError{
		Code:    "AUTH_INVALID_TOKEN_CLAIM",
		Message: "token claim不合法",
	}
	ErrRoleNotFount = &BusinessError{
		Code:    "AUTH_ROLE_NOT_FOUND",
		Message: "角色不存在",
	}
	ErrUserUnauthorized = &BusinessError{ //未通过身份认证
		Code:    "AUTH_UNAUTHORIZED",
		Message: "用户未授权",
	}
	ErrPermissionDenied = &BusinessError{ //越权
		Code:    "AUTH_PERMISSION_DENIED",
		Message: "权限不足",
	}
	// 用户侧：
	ErrUserInvalidCredentials = &BusinessError{
		Code:    "USER_INVALID_CREDENTIALS", //无效凭据
		Message: "用户名或密码错误",
	}
	ErrInvalidUserID = &BusinessError{
		Code:    "USER_INVALID_ID",
		Message: "无效的用户ID",
	}
	ErrUserNotFound = &BusinessError{
		Code:    "USER_NOT_FOUND",
		Message: "找不到该用户",
	}
	ErrUserAlreadyExists = &BusinessError{
		Code:    "USER_ALREADY_EXISTS",
		Message: "该用户名已经存在",
	}
	ErrInvalidUserRole = &BusinessError{ //越权
		Code:    "USER_INVALID_ROLE",
		Message: "角色不合法",
	}
	ErrInvalidPassword = &BusinessError{
		Code:    "USER_INVALID_PASSWORD",
		Message: "密码错误",
	}
	ErrPasswordTooShort = &BusinessError{
		Code:    "USER_PASSWORD_TOO_SHORT",
		Message: "密码长度不足 8 位",
	}

	// 订单侧：
	ErrInvalidOrderID = &BusinessError{
		Code:    "ORDER_INVALID_ID",
		Message: "无效的订单ID",
	}
	ErrOrderNotFound = &BusinessError{
		Code:    "ORDER_NOT_FOUND",
		Message: "找不到该订单",
	}
	ErrRoleSideMismatch = &BusinessError{
		Code:    "ORDER_ROLE_SIDE_MISMATCH",
		Message: "角色与订单方向不匹配",
	}
	ErrOrderNotCancellable = &BusinessError{
		Code:    "ORDER_NOT_CANCELLABLE",
		Message: "订单无法取消",
	}
	ErrOrderOverFilled = &BusinessError{
		Code:    "ORDER_OVER_FILLED",
		Message: "订单已超额",
	}
	ErrOrderUpdateFailed = &BusinessError{
		Code:    "ORDER_UPDATE_FAILED",
		Message: "订单更新失败",
	}
	ErrOrderLimitWithoutPrice = &BusinessError{
		Code:    "ORDER_LIMIT_WITHOUT_PRICE",
		Message: "限价单需要价格",
	}
	ErrOrderMarketWithPrice = &BusinessError{
		Code:    "ORDER_MARKET_WITH_PRICE",
		Message: "市价单不能有价格",
	}
	ErrInvalidOrderQuery = &BusinessError{
		Code:    "ORDER_INVALID_QUERY",
		Message: "无效的订单查询",
	}
	ErrInvalidOrderStatusQuery = &BusinessError{ //
		Code:    "ORDER_INVALID_STATUS_QUERY",
		Message: "无效的订单状态查询",
	}
	ErrOrderInvalidType = &BusinessError{
		Code:    "ORDER_INVALID_TYPE",
		Message: "无效的订单类型",
	}
)
