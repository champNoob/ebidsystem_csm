package handler

import (
	"ebidsystem_csm/internal/api/dto/response"
	"ebidsystem_csm/internal/apperror"

	"github.com/gin-gonic/gin"
)

func respondError(c *gin.Context, err error) {
	c.JSON(apperror.HTTPStatusOf(err), response.ErrorResponse{
		Code:    apperror.CodeOf(err),
		Message: apperror.MessageOf(err),
	})
}
