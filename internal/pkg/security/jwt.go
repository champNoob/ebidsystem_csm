package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("CHANGE_ME_IN_PROD") // TODO: 后期从 config 注入

func JWTSecret() []byte {
	return jwtSecret
}

// 自定义 JWT Claims
type CustomClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// 生成访问令牌
func GenerateJWT(userID int64, role string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
