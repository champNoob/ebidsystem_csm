package auth

import (
	"ebidsystem_csm/internal/service"
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
					"code":    service.ErrRoleNotFount.Code,
					"message": service.ErrRoleNotFount.Message,
				},
			)
			return
		}

		role, ok := roleVal.(string)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"code":    service.ErrInvalidUserRole.Code,
					"message": service.ErrInvalidUserRole.Message,
				},
			)
			return
		}

		if _, ok := roleSet[role]; !ok {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"code":    service.ErrPermissionDenied.Code,
					"message": service.ErrPermissionDenied.Message,
				},
			)
			return
		}

		c.Next()
	}
}
