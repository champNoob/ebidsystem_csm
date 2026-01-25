package auth

import (
	"net/http"
	"strings"

	"ebidsystem_csm/internal/pkg/security"
	"ebidsystem_csm/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware：登录态校验
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. 从 Header 取 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"code":    service.ErrMissingAuthHeader.Code,
					"message": service.ErrMissingAuthHeader.Message,
				},
			)
			return
		}

		// 2. Bearer Token 格式校验
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"code":    service.ErrInvalidAuthHeader.Code,
					"message": service.ErrInvalidAuthHeader.Message,
				},
			)
			return
		}

		tokenStr := parts[1]

		// 3. 解析 JWT
		token, err := jwt.ParseWithClaims(
			tokenStr,
			&security.CustomClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return security.JWTSecret(), nil
			},
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"code":    service.ErrInvalidToken.Code,
					"message": service.ErrInvalidToken.Message,
				},
			)
			return
		}

		claims, ok := token.Claims.(*security.CustomClaims)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"code":    service.ErrInvalidTokenClaims.Code,
					"message": service.ErrInvalidTokenClaims.Message,
				},
			)
			return
		}

		// 4. 把用户信息塞进 Gin Context
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}
