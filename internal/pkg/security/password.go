package security

import "golang.org/x/crypto/bcrypt"

// 将明文密码安全哈希加密
func HashPassword(plain string) (string, error) {
	// bcrypt.DefaultCost 在安全性与性能间是合理折中
	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(plain),
		bcrypt.DefaultCost,
	)
	return string(bytes), err
}

// 校验明文密码与哈希是否匹配
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
