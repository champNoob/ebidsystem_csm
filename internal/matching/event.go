package matching

// 一次成功的撮合成交
type MatchEvent struct {
	EventID string
	// Symbol      string
	BuyOrderID  uint64
	SellOrderID uint64
	Price       float64
	Quantity    int64
}
