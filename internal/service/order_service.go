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
		Status:   "pending",
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
