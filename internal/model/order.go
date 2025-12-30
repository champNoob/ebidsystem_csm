package model

import "time"

// Order 订单模型（最小可用版本）
type Order struct {
	ID             int64       `json:"id"`
	UserID         int64       `json:"user_id"`         // 下单用户
	Symbol         string      `json:"symbol"`          // 证券代码，如 AAPL
	Type           OrderType   `json:"type"`            // limit / market
	Side           OrderSide   `json:"side"`            // buy / sell
	Price          *float64    `json:"price"`           // 价格
	Quantity       int64       `json:"quantity"`        // 数量
	FilledQuantity int64       `json:"filled_quantity"` // 已成交数量
	Status         OrderStatus `json:"status"`          // pending（默认） / canceled / filled / partial
	CreatedAt      time.Time   `json:"created_at"`
}
