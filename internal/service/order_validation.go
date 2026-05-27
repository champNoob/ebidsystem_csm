package service

import (
	"ebidsystem_csm/internal/apperror"
	"ebidsystem_csm/internal/model"
)

func validateRoleSide(role model.UserRole, side model.OrderSide) error {
	switch role {
	case model.UserRoleClient:
		if side != model.OrderSideBuy {
			return apperror.ErrRoleSideMismatch
		}
	case model.UserRoleSeller:
		if side != model.OrderSideSell {
			return apperror.ErrRoleSideMismatch
		}
	case model.UserRoleTrader:
		return nil
	default:
		return apperror.ErrInvalidUserRole
	}
	return nil
}
