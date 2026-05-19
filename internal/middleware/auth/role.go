package auth

import (
	"ebidsystem_csm/internal/apperror"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 角色鉴权
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{})
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"code":    apperror.ErrRoleNotFound.Code,
					"message": apperror.ErrRoleNotFound.Message,
				},
			)
			return
		}

		role, ok := roleVal.(string)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"code":    apperror.ErrInvalidUserRole.Code,
					"message": apperror.ErrInvalidUserRole.Message,
				},
			)
			return
		}

		if _, ok := roleSet[role]; !ok {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"code":    apperror.ErrPermissionDenied.Code,
					"message": apperror.ErrPermissionDenied.Message,
				},
			)
			return
		}

		c.Next()
	}
}
