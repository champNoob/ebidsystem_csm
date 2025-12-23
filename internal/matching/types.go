package matching

import "time"

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

type Order struct {
	ID        uint64
	UserID    uint64
	Symbol    string
	Side      Side
	Price     float64
	Quantity  int64
	CreatedAt time.Time
}
