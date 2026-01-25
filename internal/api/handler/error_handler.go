package handler

import (
	"net/http"

	"ebidsystem_csm/internal/api/dto/response"
	"ebidsystem_csm/internal/service"

	"github.com/gin-gonic/gin"
)

// 错误码 → HTTP 状态码映射表：
var errorCodeToHTTPStatus = map[string]int{
	"INVALID_INPUT":       http.StatusBadRequest,          // 400
	"PERMISSION_DENIED":   http.StatusForbidden,           // 403
	"ROLE_SIDE_MISMATCH":  http.StatusForbidden,           // 403
	"USER_NOT_FOUND":      http.StatusNotFound,            // 404
	"USER_ALREADY_EXISTS": http.StatusConflict,            // 409
	"INTERNAL_ERROR":      http.StatusInternalServerError, // 500
}

func respondError(c *gin.Context, err error) {
	// 业务错误
	if be, ok := err.(*service.BusinessError); ok {
		status, exists := errorCodeToHTTPStatus[be.Code]
		if !exists {
			status = http.StatusBadRequest
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
