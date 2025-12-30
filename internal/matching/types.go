package matching

import (
	"time"
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type Order struct {
	ID        uint64
	UserID    uint64
	Symbol    string
	Type      OrderType
	Side      OrderSide
	Price     float64
	Quantity  int64 //原始下单量（只读）
	Remaining int64 //剩余可成交量（会变化）
	CreatedAt time.Time
}
