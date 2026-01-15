package model

import "fmt"

type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleClient UserRole = "client"
	UserRoleSeller UserRole = "seller"
	UserRoleSales  UserRole = "sales"
	UserRoleTrader UserRole = "trader"
)

func ParseUserRole(s string) (UserRole, error) {
	switch s {
	case string(UserRoleAdmin):
		return UserRoleAdmin, nil
	case string(UserRoleClient):
		return UserRoleClient, nil
	case string(UserRoleSeller):
		return UserRoleSeller, nil
	case string(UserRoleSales):
		return UserRoleSales, nil
	case string(UserRoleTrader):
		return UserRoleTrader, nil
	default:
		return "", fmt.Errorf("invalid user role: %s", s)
	}
}
