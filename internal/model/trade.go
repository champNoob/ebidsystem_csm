package model

import "time"

type Trade struct {
	ID      uint64
	EventID string
	Symbol  string // 冗余字段，用于快速查询（Order 中已包含）
	// 不直接关联 user_id ，因为 Order 已经包含
	BuyOrderID  uint64
	SellOrderID uint64
	Price       float64
	Quantity    int64
	CreatedAt   time.Time
}
