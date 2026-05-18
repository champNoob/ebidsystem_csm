package dto

type SymbolStat struct {
	Symbol   string  `json:"symbol"`
	Trades   int64   `json:"trades"`
	Volume   int64   `json:"volume"`
	Turnover float64 `json:"turnover"`
}

type OrderStatusStat struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type UserRoleStat struct {
	Role  string `json:"role"`
	Count int64  `json:"count"`
}

type RecentTrade struct {
	ID          uint64  `json:"id"`
	EventID     string  `json:"event_id"`
	Symbol      string  `json:"symbol"`
	BuyOrderID  uint64  `json:"buy_order_id"`
	SellOrderID uint64  `json:"sell_order_id"`
	Price       float64 `json:"price"`
	Quantity    int64   `json:"quantity"`
	CreatedAt   string  `json:"created_at"`
}

type TradeTimelinePoint struct {
	TimeBucket string  `json:"time_bucket"`
	Trades     int64   `json:"trades"`
	Volume     int64   `json:"volume"`
	Turnover   float64 `json:"turnover"`
}

type UserRank struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	Volume int64  `json:"volume"`
}
