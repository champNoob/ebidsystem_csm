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
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	BuyVolume   int64  `json:"buy_volume"`
	SellVolume  int64  `json:"sell_volume"`
	TotalVolume int64  `json:"total_volume"`
}

/* 订单管理 */

type AdminOrderQuery struct {
	UserID   *int64
	Symbol   string
	Status   string
	Side     string
	Type     string
	Page     int
	PageSize int
}

type AdminOrderItem struct {
	ID             uint64   `json:"id"`
	UserID         uint64   `json:"user_id"`
	Symbol         string   `json:"symbol"`
	Type           string   `json:"type"`
	Side           string   `json:"side"`
	Price          *float64 `json:"price"`
	Quantity       int64    `json:"quantity"`
	FilledQuantity int64    `json:"filled_quantity"`
	Status         string   `json:"status"`
	CreatedAt      string   `json:"created_at"`
}

type AdminOrderPage struct {
	Items    []AdminOrderItem `json:"items"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	Total    int64            `json:"total"`
}
