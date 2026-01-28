package handler

import "ebidsystem_csm/internal/service"

var ( //只适合放 handler 自己构造的临时错误 或 HTTP 层语法错误（例如 query 解析失败）
	ErrMissingAuthHeader = &service.BusinessError{
		Code:    "AUTH_HEADER_MISSING",
		Message: "未登录",
	}

	ErrInvalidToken = &service.BusinessError{
		Code:    "TOKEN_INVALID",
		Message: "登录已失效",
	}
)
