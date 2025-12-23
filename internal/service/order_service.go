package service

import (
	"context"
	"ebidsystem_csm/internal/model"
	"ebidsystem_csm/internal/repository"
	"errors"
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

// CreateOrder 下单
func (s *OrderService) CreateOrder(
	ctx context.Context,
	userID int64,
	symbol, side string,
	price float64,
	quantity int64,
) error {

	order := &model.Order{
		UserID:   userID,
		Symbol:   symbol,
		Side:     side,
		Price:    price,
		Quantity: quantity,
		Status:   model.OrderStatusPending,
	}

	return s.repo.Create(ctx, order)
}

// ListOrders 查询订单
func (s *OrderService) ListOrders(
	ctx context.Context,
	userID int64,
	role string,
) ([]*model.Order, error) {

	if role == "admin" || role == "trader" {
		return s.repo.FindAll(ctx)
	}

	if role == "client" || role == "seller" {
		return s.repo.FindByUserID(ctx, userID)
	}

	return nil, errors.New("unauthorized role")
}

func (s *OrderService) CancelOrder(
	ctx context.Context,
	orderID int64,
	userID int64,
	role string,
) error {

	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return ErrOrderNotFound
	}

	// 1. 状态校验
	if !order.Status.CanCancel() { // 订单强类型
		return ErrOrderNotCancelable
	}

	// 2. 权限校验
	if role != "admin" && order.UserID != userID {
		return ErrPermissionDenied
	}

	// 3. 状态更新
	return s.repo.UpdateStatus(ctx, orderID, string(model.OrderStatusCanceled))
}
