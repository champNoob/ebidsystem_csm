package request

import "ebidsystem_csm/internal/model"

// 下单请求
type CreateOrderRequest struct {
	Symbol    string          `json:"symbol" binding:"required"`
	OrderType model.OrderType `json:"type"`
	OrderSide model.OrderSide `json:"side" binding:"required,oneof=buy sell"`
	Price     *float64        `json:"price" binding:"required,gt=0"`
	Quantity  int64           `json:"quantity" binding:"required,gt=0"`
}
