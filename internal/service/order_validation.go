package service

import (
	"ebidsystem_csm/internal/model"
	"errors"
)

func validateRoleSide(role model.UserRole, side model.OrderSide) error {
	switch role {
	case model.UserRoleClient:
		if side != model.OrderSideBuy {
			return errors.New("client is only allowed to place buy orders")
		}
	case model.UserRoleSeller:
		if side != model.OrderSideSell {
			return errors.New("seller is only allowed to place sell orders")
		}
	case model.UserRoleTrader:
		return nil
	default:
		return errors.New("invalid user role")
	}
	return nil
}
