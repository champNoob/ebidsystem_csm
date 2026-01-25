package service

import (
	"ebidsystem_csm/internal/model"
)

func validateRoleSide(role model.UserRole, side model.OrderSide) error {
	switch role {
	case model.UserRoleClient:
		if side != model.OrderSideBuy {
			return ErrRoleSideMismatch
		}
	case model.UserRoleSeller:
		if side != model.OrderSideSell {
			return ErrRoleSideMismatch
		}
	case model.UserRoleTrader:
		return nil
	default:
		return ErrInvalidUserRole
	}
	return nil
}
