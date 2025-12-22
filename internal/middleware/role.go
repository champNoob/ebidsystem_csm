package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware：角色鉴权
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{})
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{"error": "role not found"},
			)
			return
		}

		role, ok := roleVal.(string)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{"error": "invalid role type"},
			)
			return
		}

		if _, ok := roleSet[role]; !ok {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{"error": "permission denied"},
			)
			return
		}

		c.Next()
	}
}
