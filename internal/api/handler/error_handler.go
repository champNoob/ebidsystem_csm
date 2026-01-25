package handler

import (
	"net/http"

	"ebidsystem_csm/internal/api/dto/response"
	"ebidsystem_csm/internal/service"

	"github.com/gin-gonic/gin"
)

// 错误码 → HTTP 状态码映射表：
var errorCodeToHTTPStatus = map[string]int{
	// 通用
	"INVALID_INPUT":  http.StatusBadRequest,          // 400
	"INTERNAL_ERROR": http.StatusInternalServerError, // 500
	// AUTH 域
	"AUTH_MISSING_HEADER":      http.StatusUnauthorized, // 401
	"AUTH_INVALID_HEADER":      http.StatusUnauthorized, // 401
	"AUTH_INVALID_TOKEN":       http.StatusUnauthorized, // 401
	"AUTH_INVALID_TOKEN_CLAIM": http.StatusUnauthorized, // 401
	"AUTH_UNAUTHORIZED":        http.StatusUnauthorized, // 401
	"AUTH_PERMISSION_DENIED":   http.StatusForbidden,    // 403
	"AUTH_ROLE_NOT_FOUND":      http.StatusForbidden,    // 403
	// USER 域
	"USER_INVALID_CREDENTIALS": http.StatusUnauthorized, // 401
	"USER_INVALID_ID":          http.StatusBadRequest,   // 400
	"USER_NOT_FOUND":           http.StatusNotFound,     // 404
	"USER_ALREADY_EXISTS":      http.StatusConflict,     // 409
	"USER_INVALID_ROLE":        http.StatusBadRequest,   // 400
	"USER_INVALID_PASSWORD":    http.StatusUnauthorized, // 401
	"USER_PASSWORD_TOO_SHORT":  http.StatusBadRequest,   // 400
	// ORDER 域
	"ORDER_INVALID_ID":           http.StatusBadRequest, // 400
	"ORDER_NOT_FOUND":            http.StatusNotFound,   // 404
	"ORDER_ROLE_SIDE_MISMATCH":   http.StatusForbidden,  // 403
	"ORDER_NOT_CANCELLABLE":      http.StatusConflict,   // 409
	"ORDER_OVER_FILLED":          http.StatusConflict,   // 409
	"ORDER_UPDATE_FAILED":        http.StatusConflict,   // 409
	"ORDER_LIMIT_WITHOUT_PRICE":  http.StatusBadRequest, // 400
	"ORDER_MARKET_WITH_PRICE":    http.StatusBadRequest, // 400
	"ORDER_INVALID_QUERY":        http.StatusBadRequest, // 400
	"ORDER_INVALID_STATUS_QUERY": http.StatusBadRequest, // 400
	"ORDER_INVALID_TYPE":         http.StatusBadRequest, // 400
}

func respondError(c *gin.Context, err error) {
	// 业务错误
	if be, ok := err.(*service.BusinessError); ok {
		status, exists := errorCodeToHTTPStatus[be.Code]
		if !exists {
			status = http.StatusInternalServerError
		}

		c.JSON(status, response.ErrorResponse{
			Code:    be.Code,
			Message: be.Message,
		})
		return
	}

	// 未知错误（兜底）
	c.JSON(http.StatusInternalServerError, response.ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "内部错误，请稍后重试",
	})
}
